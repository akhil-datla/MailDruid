package scheduler

import (
	"context"
	"fmt"
	"log"
	"main/components/platform/postgresmanager"
	"main/components/user"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type tasklist map[string][]string

type TaskManager struct {
	Mutex    sync.Mutex
	Tasklist tasklist
}

var Taskmanager = &TaskManager{Tasklist: make(tasklist)}

var ctx, cancel = context.WithCancel(context.Background())

func Cleanup() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Stopping processes....")
		cancel()
		os.Exit(1)
	}()

}

func ScheduleTasks() {
	Taskmanager.Mutex.Lock()
	userlist := make([]*user.User, 0)
	postgresmanager.ReadAll(&userlist)
	for _, u := range userlist {
		if u.UpdateInterval != "0" && u.UpdateInterval != "" {
			Taskmanager.Tasklist[u.UpdateInterval] = []string{u.ID}
			Schedule(u.UpdateInterval, ctx)
		}
	}
	Taskmanager.Mutex.Unlock()
}

func ScheduleNewTask(id, interval string) error {
	u, err := user.ReadUser(id)
	if err != nil {
		return err
	}

	if u.UpdateInterval != "0" && u.UpdateInterval != "" {
		return fmt.Errorf("user already has a task scheduled")
	}

	Taskmanager.Mutex.Lock()
	if _, ok := Taskmanager.Tasklist[interval]; ok {
		Taskmanager.Tasklist[interval] = append(Taskmanager.Tasklist[interval], u.ID)
		postgresmanager.Update(&u, &user.User{UpdateInterval: interval})
	} else {
		Taskmanager.Tasklist[interval] = []string{u.ID}
		postgresmanager.Update(&u, &user.User{UpdateInterval: interval})
		Schedule(interval, ctx)
	}
	Taskmanager.Mutex.Unlock()
	return nil
}

func DeleteTask(id, interval string) error {
	u, err := user.ReadUser(id)
	if err != nil {
		return err
	}
	Taskmanager.Mutex.Lock()
	if _, ok := Taskmanager.Tasklist[interval]; ok {
		for i, uID := range Taskmanager.Tasklist[interval] {
			if uID == id {
				Taskmanager.Tasklist[interval] = remove(Taskmanager.Tasklist[interval], i)
			}
		}
	} else {
		return fmt.Errorf("no task with interval %s", interval)
	}

	if len(Taskmanager.Tasklist[interval]) == 0 {
		cancel()
		delete(Taskmanager.Tasklist, interval)
	}

	Taskmanager.Mutex.Unlock()
	postgresmanager.Update(&u, &user.User{UpdateInterval: "0"})
	return nil

}

func DeleteTaskforUser(id string) error {
	u, err := user.ReadUser(id)
	if err != nil {
		return err
	}
	Taskmanager.Mutex.Lock()
	if u.UpdateInterval != "" {
		for i, uID := range Taskmanager.Tasklist[u.UpdateInterval] {
			if uID == id {
				Taskmanager.Tasklist[u.UpdateInterval] = remove(Taskmanager.Tasklist[u.UpdateInterval], i)
			}
		}
	}

	if len(Taskmanager.Tasklist[u.UpdateInterval]) == 0 {
		cancel()
		delete(Taskmanager.Tasklist, u.UpdateInterval)
	}

	Taskmanager.Mutex.Unlock()
	postgresmanager.Update(&u, &user.User{UpdateInterval: "0"})
	return nil
}

func UpdateTask(id, oldInterval, newInterval string) error {
	err := DeleteTask(id, oldInterval)
	if err != nil {
		Taskmanager.Mutex.Unlock()
		return err
	}
	err = ScheduleNewTask(id, newInterval)
	if err != nil {
		Taskmanager.Mutex.Unlock()
		return err
	}
	return nil
}

func remove(s []string, i int) []string {
	if s == nil {
		return nil
	}
	if i >= len(s) {
		return s
	}
	s = append(s[:i], s[i+1:]...)
	return s
}

func Schedule(interval string, ctx context.Context) {
	intervalInt, err := strconv.Atoi(interval)
	if err != nil {
		log.Println(err)
	}

	ticker := time.NewTicker(time.Duration(intervalInt) * time.Minute)
	go func(interval string, ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				Taskmanager.Mutex.Lock()
				for _, id := range Taskmanager.Tasklist[interval] {
					user, err := user.ReadUser(id)
					if err != nil {
						log.Println(err)
					}
					summary, fileName, err := user.GenerateSummaryandWordCloud()
					if err == nil {
						err = user.SendEmail(summary, fileName, "")
						if err != nil {
							log.Println(err)
						}
						err := os.Remove(fileName)
						if err != nil {
							log.Println(err)
						}
					} else if err.Error() == fmt.Sprintf("no emails found with tags: %s", user.Tags) {
						err1 := user.SendEmail("", "", "No new emails to summarize.")
						if err != nil {
							log.Println(err1)
						}
					} else {
						err2 := user.SendEmail("", "", fmt.Sprintf("Error: %s", err.Error()))
						if err != nil {
							log.Println(err2)
						}
					}
				}
				Taskmanager.Mutex.Unlock()
			}
		}
	}(interval, ctx)
}
