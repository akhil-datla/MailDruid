package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"main/components/platform/email"
	"main/components/platform/encryption"
	"main/components/scheduler"
	"main/components/user"
	"main/confidential"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type jwtCustomClaims struct {
	ID string
	jwt.StandardClaims
}

func createUser(c echo.Context) error {
	if containsEmpty(c.FormValue("name"), c.FormValue("email"), c.FormValue("receivingEmail"), c.FormValue("password"), c.FormValue("domain"), c.FormValue("port")) {
		return c.JSON(http.StatusBadRequest, "Empty fields are not allowed.")
	}

	port, err := strconv.Atoi(c.FormValue("port"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Port must be a number.")
	}

	err = user.CreateUser(c.FormValue("name"), c.FormValue("email"), c.FormValue("receivingEmail"), c.FormValue("password"), c.FormValue("domain"), port)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "User created successfully.")
}

func loginUser(c echo.Context) error {
	if containsEmpty(c.FormValue("email"), c.FormValue("password")) {
		return c.JSON(http.StatusBadRequest, "Empty fields are not allowed.")
	}

	id, err := user.AuthUser(c.FormValue("email"), c.FormValue("password"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	sclaims := jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	}

	claims := &jwtCustomClaims{
		ID:             id,
		StandardClaims: sclaims,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString(confidential.SigningKey)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not generate token.")
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
	})

}

func getUser(c echo.Context) error {
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(*jwtCustomClaims)
	id := claims.ID
	user, err := user.ReadUser(id)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	user.ID = ""
	user.Password = ""
	user.LastUID = ""

	return c.JSON(http.StatusOK, user)
}

func updateUser(c echo.Context) error {

	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(*jwtCustomClaims)
	id := claims.ID

	port, err := strconv.Atoi(c.FormValue("port"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Port must be a number.")
	}

	errors := user.UpdateUser(id, c.FormValue("name"), c.FormValue("email"), c.FormValue("receivingEmail"), c.FormValue("oldPassword"), c.FormValue("newPassword"), c.FormValue("domain"), c.FormValue("folder"), port)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors)
	}
	return c.String(http.StatusOK, "User updated successfully.")
}

func deleteUser(c echo.Context) error {
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(*jwtCustomClaims)
	id := claims.ID

	err := scheduler.DeleteTaskforUser(id)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	err = user.DeleteUser(id)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, "User deleted successfully.")
}

func getFolders(c echo.Context) error {
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(*jwtCustomClaims)
	id := claims.ID

	user, err := user.ReadUser(id)

	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	rawPass, err := base64.RawStdEncoding.DecodeString(user.Password)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not decode password.")
	}
	im, err := email.Init(user.Email, string(encryption.Decrypt(rawPass)), user.Domain, user.Port)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not initialize email client.")
	}

	folders, err := im.GetFolders()
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, folders)
}

func updateFolder(c echo.Context) error {
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(*jwtCustomClaims)
	id := claims.ID

	err := user.UpdateFolder(id, c.FormValue("folder"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Folder updated successfully.")
}

func updateTags(c echo.Context) error {
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(*jwtCustomClaims)
	id := claims.ID

	if containsEmpty(c.FormValue("tags")) {
		return c.String(http.StatusBadRequest, "Must have at least one tag.")
	}

	tags := []string{}
	err := json.Unmarshal([]byte(c.FormValue("tags")), &tags)
	if err != nil {
		return c.String(http.StatusBadRequest, "Tags must be a string array.")
	}

	err = user.UpdateTags(id, tags)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Tags updated successfully.")
}

func updateSummaryCount(c echo.Context) error {
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(*jwtCustomClaims)
	id := claims.ID

	if containsEmpty(c.FormValue("summaryCount")) {
		return c.String(http.StatusBadRequest, "Cannot be empty.")
	}

	summaryCount, err := strconv.Atoi(c.FormValue("summaryCount"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Must be a number.")
	}

	if summaryCount < 1 {
		return c.String(http.StatusBadRequest, "Must be greater than 0.")
	}

	err = user.UpdateSummaryCount(id, summaryCount)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Summary count updated successfully.")
}

func newTask(c echo.Context) error {
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(*jwtCustomClaims)
	id := claims.ID

	err := scheduler.ScheduleNewTask(id, c.FormValue("interval"))

	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, "Task scheduled successfully.")
}

func updateTask(c echo.Context) error {
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(*jwtCustomClaims)
	id := claims.ID

	err := scheduler.UpdateTask(id, c.FormValue("oldInterval"), c.FormValue("newInterval"))

	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, "Task updated successfully.")
}

func deleteTask(c echo.Context) error {
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(*jwtCustomClaims)
	id := claims.ID

	err := scheduler.DeleteTask(id, c.FormValue("interval"))

	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, "Task deleted successfully.")
}

func getTaskList(c echo.Context) error {
	return c.JSON(http.StatusOK, scheduler.Taskmanager.Tasklist)
}

func generateSummaryandWordCloud(c echo.Context) error {
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(*jwtCustomClaims)
	id := claims.ID

	user, err := user.ReadUser(id)

	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	summary, fileName, err := user.GenerateSummaryandWordCloud()
	if err != nil && err.Error() == fmt.Sprintf("no emails found with tags: %s", user.Tags) {
		return c.String(http.StatusOK, "No new emails to summarize.")
	} else if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	var base64Image string

	if fileName != "" {
		pathToFile := filepath.Join(filepath.Dir(""), fileName)

		bytes, err := os.ReadFile(pathToFile)

		if err != nil {
			return c.String(http.StatusInternalServerError, "Could not read file.")
		}

		base64Image = base64.RawStdEncoding.EncodeToString(bytes)

		removeErr := os.Remove(pathToFile)
		if removeErr != nil {
			log.Println(removeErr)
		}
	}

	m := map[string]interface{}{"summary": summary, "image": base64Image}

	return c.JSON(http.StatusOK, m)
}

func containsEmpty(ss ...string) bool {
	for _, s := range ss {
		if s == "" {
			return true
		}
	}
	return false
}
