const API_BASE = '/api/v1';

class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, body.error || body.message || 'Request failed');
  }

  return res.json();
}

// Auth
export const register = (data: {
  name: string;
  email: string;
  receivingEmail: string;
  password: string;
  domain: string;
  port: number;
}) => request<{ message: string }>('/users', { method: 'POST', body: JSON.stringify(data) });

export const login = (email: string, password: string) =>
  request<{ token: string }>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });

// User
export const getProfile = () => request<UserProfile>('/users/me');

export const updateProfile = (data: Record<string, unknown>) =>
  request<{ message: string }>('/users/me', { method: 'PATCH', body: JSON.stringify(data) });

export const deleteAccount = () =>
  request<{ message: string }>('/users/me', { method: 'DELETE' });

// Email config
export const getFolders = () => request<string[]>('/users/me/folders');

export const updateFolder = (folder: string) =>
  request<{ message: string }>('/users/me/folder', {
    method: 'PATCH',
    body: JSON.stringify({ folder }),
  });

export const updateTags = (tags: string[]) =>
  request<{ message: string }>('/users/me/tags', {
    method: 'PUT',
    body: JSON.stringify({ tags }),
  });

export const updateBlacklist = (senders: string[]) =>
  request<{ message: string }>('/users/me/blacklist', {
    method: 'PUT',
    body: JSON.stringify({ senders }),
  });

export const updateStartTime = (startTime: string) =>
  request<{ message: string }>('/users/me/start-time', {
    method: 'PATCH',
    body: JSON.stringify({ startTime }),
  });

export const updateSummaryCount = (count: number) =>
  request<{ message: string }>('/users/me/summary-count', {
    method: 'PATCH',
    body: JSON.stringify({ count }),
  });

// Schedules
export const getSchedules = () => request<TaskInfo[]>('/schedules');

export const createSchedule = (interval: string) =>
  request<{ message: string }>('/schedules', {
    method: 'POST',
    body: JSON.stringify({ interval }),
  });

export const updateSchedule = (oldInterval: string, newInterval: string) =>
  request<{ message: string }>('/schedules', {
    method: 'PATCH',
    body: JSON.stringify({ oldInterval, newInterval }),
  });

export const deleteSchedule = (interval: string) =>
  request<{ message: string }>('/schedules', {
    method: 'DELETE',
    body: JSON.stringify({ interval }),
  });

// Summary
export const generateSummary = () =>
  request<SummaryResult>('/summaries/generate', { method: 'POST' });

// Types
export interface UserProfile {
  id: string;
  name: string;
  email: string;
  receivingEmail: string;
  domain: string;
  port: number;
  folder: string;
  tags: string[] | null;
  blackListSenders: string[] | null;
  startTime: string;
  summaryCount: number;
  updateInterval: string;
}

export interface TaskInfo {
  interval: string;
  userIds: string[];
}

export interface SummaryResult {
  summary: string;
  image?: string;
}

export { ApiError };
