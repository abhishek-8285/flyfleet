export interface DropdownOption {
  value: string;
  label: string;
  icon?: string;
  description?: string;
}

export interface ApiResponse<T> {
  data: T;
  success: boolean;
  message?: string;
  error?: string;
}
