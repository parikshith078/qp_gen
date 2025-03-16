// User data type
export interface UserModel {
  id: string;
  name: string;
  email: string;
  username: string;
  created_at: string;
  updated_at: string;
  last_activity: string;
}

// API response type for user operations
export interface RegisterUserResponse {
  error: boolean;
  message: string;
  data: UserModel;
}

// Generic API response type
export interface ApiResponse<T = any> {
  error: boolean;
  message: string;
  data?: T;
}
