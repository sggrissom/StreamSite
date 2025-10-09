import * as rpc from "vlens/rpc";

export type UserRole = number;
export const RoleUser: UserRole = 0;
export const RoleStreamAdmin: UserRole = 1;
export const RoleSiteAdmin: UserRole = 2;

// Errors
export const ErrLoginFailure = "LoginFailure";
export const ErrAuthFailure = "AuthFailure";

export interface CreateAccountRequest {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
}

export interface CreateAccountResponse {
  success: boolean;
  error: string;
  token: string;
  auth: AuthResponse;
}

export interface Empty {}

export interface AuthResponse {
  id: number;
  name: string;
  email: string;
  role: UserRole;
  isStreamAdmin: boolean;
  isSiteAdmin: boolean;
}

export interface SetUserRoleRequest {
  userId: number;
  role: UserRole;
}

export interface SetUserRoleResponse {
  success: boolean;
  error: string;
}

export interface ListUsersRequest {}

export interface ListUsersResponse {
  users: UserListInfo[];
}

export interface UserListInfo {
  id: number;
  name: string;
  email: string;
  role: UserRole;
  roleName: string;
}

export async function CreateAccount(
  data: CreateAccountRequest,
): Promise<rpc.Response<CreateAccountResponse>> {
  return await rpc.call<CreateAccountResponse>(
    "CreateAccount",
    JSON.stringify(data),
  );
}

export async function GetAuthContext(
  data: Empty,
): Promise<rpc.Response<AuthResponse>> {
  return await rpc.call<AuthResponse>("GetAuthContext", JSON.stringify(data));
}

export async function SetUserRole(
  data: SetUserRoleRequest,
): Promise<rpc.Response<SetUserRoleResponse>> {
  return await rpc.call<SetUserRoleResponse>(
    "SetUserRole",
    JSON.stringify(data),
  );
}

export async function ListUsers(
  data: ListUsersRequest,
): Promise<rpc.Response<ListUsersResponse>> {
  return await rpc.call<ListUsersResponse>("ListUsers", JSON.stringify(data));
}
