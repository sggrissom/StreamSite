import * as rpc from "vlens/rpc";

export type UserRole = number;
export const RoleUser: UserRole = 0;
export const RoleStreamAdmin: UserRole = 1;
export const RoleSiteAdmin: UserRole = 2;

export type StudioRole = number;
export const StudioRoleViewer: StudioRole = 0;
export const StudioRoleMember: StudioRole = 1;
export const StudioRoleAdmin: StudioRole = 2;
export const StudioRoleOwner: StudioRole = 3;

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

export interface CreateStudioRequest {
  name: string;
  description: string;
  maxRooms: number;
}

export interface CreateStudioResponse {
  success: boolean;
  error: string;
  studio: Studio;
}

export interface ListMyStudiosRequest {}

export interface ListMyStudiosResponse {
  studios: StudioWithRole[];
}

export interface GetStudioRequest {
  studioId: number;
}

export interface GetStudioResponse {
  success: boolean;
  error: string;
  studio: Studio;
  myRole: StudioRole;
  myRoleName: string;
}

export interface GetStudioDashboardRequest {
  studioId: number;
}

export interface GetStudioDashboardResponse {
  success: boolean;
  error: string;
  studio: Studio;
  myRole: StudioRole;
  myRoleName: string;
  rooms: Room[];
}

export interface UpdateStudioRequest {
  studioId: number;
  name: string;
  description: string;
  maxRooms: number;
}

export interface UpdateStudioResponse {
  success: boolean;
  error: string;
  studio: Studio;
}

export interface DeleteStudioRequest {
  studioId: number;
}

export interface DeleteStudioResponse {
  success: boolean;
  error: string;
}

export interface CreateRoomRequest {
  studioId: number;
  name: string;
}

export interface CreateRoomResponse {
  success: boolean;
  error: string;
  room: Room;
}

export interface ListRoomsRequest {
  studioId: number;
}

export interface ListRoomsResponse {
  success: boolean;
  error: string;
  rooms: Room[];
}

export interface GetRoomStreamKeyRequest {
  roomId: number;
}

export interface GetRoomStreamKeyResponse {
  success: boolean;
  error: string;
  streamKey: string;
}

export interface UpdateRoomRequest {
  roomId: number;
  name: string;
}

export interface UpdateRoomResponse {
  success: boolean;
  error: string;
  room: Room;
}

export interface RegenerateStreamKeyRequest {
  roomId: number;
}

export interface RegenerateStreamKeyResponse {
  success: boolean;
  error: string;
  streamKey: string;
}

export interface DeleteRoomRequest {
  roomId: number;
}

export interface DeleteRoomResponse {
  success: boolean;
  error: string;
}

export interface UserListInfo {
  id: number;
  name: string;
  email: string;
  role: UserRole;
  roleName: string;
}

export interface Studio {
  id: number;
  name: string;
  description: string;
  maxRooms: number;
  ownerId: number;
  creation: string;
}

export interface StudioWithRole {
  id: number;
  name: string;
  description: string;
  maxRooms: number;
  ownerId: number;
  creation: string;
  myRole: StudioRole;
  myRoleName: string;
}

export interface Room {
  id: number;
  studioId: number;
  roomNumber: number;
  name: string;
  streamKey: string;
  isActive: boolean;
  creation: string;
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

export async function CreateStudio(
  data: CreateStudioRequest,
): Promise<rpc.Response<CreateStudioResponse>> {
  return await rpc.call<CreateStudioResponse>(
    "CreateStudio",
    JSON.stringify(data),
  );
}

export async function ListMyStudios(
  data: ListMyStudiosRequest,
): Promise<rpc.Response<ListMyStudiosResponse>> {
  return await rpc.call<ListMyStudiosResponse>(
    "ListMyStudios",
    JSON.stringify(data),
  );
}

export async function GetStudio(
  data: GetStudioRequest,
): Promise<rpc.Response<GetStudioResponse>> {
  return await rpc.call<GetStudioResponse>("GetStudio", JSON.stringify(data));
}

export async function GetStudioDashboard(
  data: GetStudioDashboardRequest,
): Promise<rpc.Response<GetStudioDashboardResponse>> {
  return await rpc.call<GetStudioDashboardResponse>(
    "GetStudioDashboard",
    JSON.stringify(data),
  );
}

export async function UpdateStudio(
  data: UpdateStudioRequest,
): Promise<rpc.Response<UpdateStudioResponse>> {
  return await rpc.call<UpdateStudioResponse>(
    "UpdateStudio",
    JSON.stringify(data),
  );
}

export async function DeleteStudio(
  data: DeleteStudioRequest,
): Promise<rpc.Response<DeleteStudioResponse>> {
  return await rpc.call<DeleteStudioResponse>(
    "DeleteStudio",
    JSON.stringify(data),
  );
}

export async function CreateRoom(
  data: CreateRoomRequest,
): Promise<rpc.Response<CreateRoomResponse>> {
  return await rpc.call<CreateRoomResponse>("CreateRoom", JSON.stringify(data));
}

export async function ListRooms(
  data: ListRoomsRequest,
): Promise<rpc.Response<ListRoomsResponse>> {
  return await rpc.call<ListRoomsResponse>("ListRooms", JSON.stringify(data));
}

export async function GetRoomStreamKey(
  data: GetRoomStreamKeyRequest,
): Promise<rpc.Response<GetRoomStreamKeyResponse>> {
  return await rpc.call<GetRoomStreamKeyResponse>(
    "GetRoomStreamKey",
    JSON.stringify(data),
  );
}

export async function UpdateRoom(
  data: UpdateRoomRequest,
): Promise<rpc.Response<UpdateRoomResponse>> {
  return await rpc.call<UpdateRoomResponse>("UpdateRoom", JSON.stringify(data));
}

export async function RegenerateStreamKey(
  data: RegenerateStreamKeyRequest,
): Promise<rpc.Response<RegenerateStreamKeyResponse>> {
  return await rpc.call<RegenerateStreamKeyResponse>(
    "RegenerateStreamKey",
    JSON.stringify(data),
  );
}

export async function DeleteRoom(
  data: DeleteRoomRequest,
): Promise<rpc.Response<DeleteRoomResponse>> {
  return await rpc.call<DeleteRoomResponse>("DeleteRoom", JSON.stringify(data));
}
