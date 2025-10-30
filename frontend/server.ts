import * as rpc from "vlens/rpc"

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
    name: string
    email: string
    password: string
    confirmPassword: string
}

export interface CreateAccountResponse {
    token: string
    auth: AuthResponse
}

export interface Empty {
}

export interface AuthResponse {
    id: number
    name: string
    email: string
    role: UserRole
    isStreamAdmin: boolean
    isSiteAdmin: boolean
    canManageStudios: boolean
}

export interface ListAllUsersRequest {
}

export interface ListAllUsersResponse {
    users: UserWithStats[]
}

export interface UpdateUserRoleRequest {
    userId: number
    newRole: UserRole
}

export interface UpdateUserRoleResponse {
    user: User
}

export interface SetUserRoleRequest {
    userId: number
    role: UserRole
}

export interface SetUserRoleResponse {
}

export interface ListUsersRequest {
}

export interface ListUsersResponse {
    users: UserListInfo[]
}

export interface CreateStudioRequest {
    name: string
    description: string
    maxRooms: number
}

export interface CreateStudioResponse {
    studio: Studio
}

export interface ListMyStudiosRequest {
}

export interface ListMyStudiosResponse {
    studios: StudioWithRole[]
}

export interface ListAllStudiosRequest {
}

export interface ListAllStudiosResponse {
    studios: StudioWithOwner[]
}

export interface GetStudioRequest {
    studioId: number
}

export interface GetStudioResponse {
    studio: Studio
    myRole: StudioRole
    myRoleName: string
}

export interface GetStudioDashboardRequest {
    studioId: number
}

export interface GetStudioDashboardResponse {
    studio: Studio
    myRole: StudioRole
    myRoleName: string
    rooms: Room[]
    members: MemberWithDetails[]
}

export interface UpdateStudioRequest {
    studioId: number
    name: string
    description: string
    maxRooms: number
}

export interface UpdateStudioResponse {
    studio: Studio
}

export interface DeleteStudioRequest {
    studioId: number
}

export interface DeleteStudioResponse {
}

export interface CreateRoomRequest {
    studioId: number
    name: string
    cameraRtsp: string | null
}

export interface CreateRoomResponse {
    room: Room
}

export interface ListRoomsRequest {
    studioId: number
}

export interface ListRoomsResponse {
    rooms: Room[]
}

export interface GetRoomDetailsRequest {
    roomId: number
}

export interface GetRoomDetailsResponse {
    room: Room
    studioName: string
    myRole: StudioRole
    myRoleName: string
    isCodeAuth: boolean
    codeExpiresAt: string | null
}

export interface GetStudioRoomsForCodeSessionRequest {
    studioId: number
}

export interface GetStudioRoomsForCodeSessionResponse {
    studioName: string
    rooms: Room[]
    codeExpiresAt: string | null
}

export interface ListMyAccessibleRoomsRequest {
}

export interface ListMyAccessibleRoomsResponse {
    rooms: RoomWithStudio[]
    codeExpiresAt: string | null
}

export interface GetRoomStreamKeyRequest {
    roomId: number
}

export interface GetRoomStreamKeyResponse {
    streamKey: string
}

export interface UpdateRoomRequest {
    roomId: number
    name: string
    cameraRtsp: string | null
}

export interface UpdateRoomResponse {
    room: Room
}

export interface RegenerateStreamKeyRequest {
    roomId: number
}

export interface RegenerateStreamKeyResponse {
    streamKey: string
}

export interface DeleteRoomRequest {
    roomId: number
}

export interface DeleteRoomResponse {
}

export interface AddStudioMemberRequest {
    studioId: number
    userId: number
    userEmail: string
    role: StudioRole
}

export interface AddStudioMemberResponse {
    membership: StudioMembership
}

export interface RemoveStudioMemberRequest {
    studioId: number
    userId: number
}

export interface RemoveStudioMemberResponse {
}

export interface UpdateStudioMemberRoleRequest {
    studioId: number
    userId: number
    newRole: StudioRole
}

export interface UpdateStudioMemberRoleResponse {
    membership: StudioMembership
}

export interface ListStudioMembersRequest {
    studioId: number
}

export interface ListStudioMembersResponse {
    members: MemberWithDetails[]
}

export interface LeaveStudioRequest {
    studioId: number
}

export interface LeaveStudioResponse {
}

export interface GenerateAccessCodeRequest {
    type: number
    targetId: number
    durationMinutes: number
    maxViewers: number
    label: string
}

export interface GenerateAccessCodeResponse {
    code: string
    expiresAt: string
    shareUrl: string
}

export interface GetCodeStreamAccessRequest {
    sessionToken: string
    roomId: number
}

export interface GetCodeStreamAccessResponse {
    allowed: boolean
    roomId: number
    studioId: number
    expiresAt: string
    gracePeriod: boolean
    message: string
}

export interface RevokeAccessCodeRequest {
    code: string
}

export interface RevokeAccessCodeResponse {
    sessionsKilled: number
}

export interface ListAccessCodesRequest {
    type: number
    targetId: number
}

export interface ListAccessCodesResponse {
    codes: AccessCodeListItem[]
}

export interface GetCodeAnalyticsRequest {
    code: string
}

export interface GetCodeAnalyticsResponse {
    code: string
    type: number
    label: string
    status: string
    createdAt: string
    expiresAt: string
    totalConnections: number
    currentViewers: number
    peakViewers: number
    peakViewersAt: string
    sessions: SessionInfo[]
}

export interface SetCameraConfigRequest {
    roomId: number
    rtspUrl: string
}

export interface SetCameraConfigResponse {
}

export interface GetCameraConfigRequest {
    roomId: number
}

export interface GetCameraConfigResponse {
    roomId: number
    rtspUrl: string | null
}

export interface DeleteCameraConfigRequest {
    roomId: number
}

export interface DeleteCameraConfigResponse {
}

export interface GetRoomAnalyticsRequest {
    roomId: number
}

export interface GetRoomAnalyticsResponse {
    analytics: RoomAnalytics | null
    isStreaming: boolean
    roomName: string
}

export interface GetStudioAnalyticsRequest {
    studioId: number
}

export interface GetStudioAnalyticsResponse {
    analytics: StudioAnalytics | null
    studioName: string
}

export interface RecalculateViewerCountsRequest {
    studioId: number
}

export interface RecalculateViewerCountsResponse {
    roomsUpdated: number
    studiosUpdated: number
    codesUpdated: number
    message: string
}

export interface SendEmoteRequest {
    roomId: number
    emote: string
}

export interface SendEmoteResponse {
}

export interface SRSAuthCallback {
    server_id: string
    action: string
    client_id: string
    ip: string
    vhost: string
    app: string
    tcUrl: string
    stream: string
    param: string
    stream_url: string
    stream_id: string
}

export interface SRSAuthResponse {
    code: number
}

export interface UserWithStats {
    id: number
    name: string
    email: string
    role: UserRole
    creation: string
    lastLogin: string
    studioCount: number
}

export interface User {
    id: number
    name: string
    email: string
    role: UserRole
    creation: string
    lastLogin: string
}

export interface UserListInfo {
    id: number
    name: string
    email: string
    role: UserRole
    roleName: string
}

export interface Studio {
    id: number
    name: string
    description: string
    maxRooms: number
    ownerId: number
    creation: string
}

export interface StudioWithRole {
    id: number
    name: string
    description: string
    maxRooms: number
    ownerId: number
    creation: string
    myRole: StudioRole
    myRoleName: string
}

export interface StudioWithOwner {
    id: number
    name: string
    description: string
    maxRooms: number
    ownerId: number
    creation: string
    ownerName: string
    ownerEmail: string
    roomCount: number
    memberCount: number
}

export interface Room {
    id: number
    studioId: number
    roomNumber: number
    name: string
    streamKey: string
    isActive: boolean
    creation: string
}

export interface MemberWithDetails {
    userId: number
    studioId: number
    role: StudioRole
    joinedAt: string
    userName: string
    userEmail: string
    roleName: string
}

export interface RoomWithStudio {
    id: number
    studioId: number
    roomNumber: number
    name: string
    streamKey: string
    isActive: boolean
    creation: string
    studioName: string
}

export interface StudioMembership {
    userId: number
    studioId: number
    role: StudioRole
    joinedAt: string
}

export interface AccessCodeListItem {
    code: string
    type: number
    label: string
    createdAt: string
    expiresAt: string
    isRevoked: boolean
    isExpired: boolean
    currentViewers: number
    totalViews: number
}

export interface SessionInfo {
    connectedAt: string
    duration: number
    clientIP: string
    userAgent: string
    isActive: boolean
}

export interface RoomAnalytics {
    roomId: number
    totalViewsAllTime: number
    totalViewsThisMonth: number
    uniqueViewersAllTime: number
    uniqueViewersThisMonth: number
    currentViewers: number
    peakViewers: number
    peakViewersAt: string
    lastStreamAt: string
    streamStartedAt: string
    totalStreamMinutes: number
}

export interface StudioAnalytics {
    studioId: number
    totalViewsAllTime: number
    totalViewsThisMonth: number
    uniqueViewersAllTime: number
    uniqueViewersThisMonth: number
    currentViewers: number
    totalRooms: number
    activeRooms: number
    totalStreamMinutes: number
}

export async function CreateAccount(data: CreateAccountRequest): Promise<rpc.Response<CreateAccountResponse>> {
    return await rpc.call<CreateAccountResponse>('CreateAccount', JSON.stringify(data));
}

export async function GetAuthContext(data: Empty): Promise<rpc.Response<AuthResponse>> {
    return await rpc.call<AuthResponse>('GetAuthContext', JSON.stringify(data));
}

export async function ListAllUsers(data: ListAllUsersRequest): Promise<rpc.Response<ListAllUsersResponse>> {
    return await rpc.call<ListAllUsersResponse>('ListAllUsers', JSON.stringify(data));
}

export async function UpdateUserRole(data: UpdateUserRoleRequest): Promise<rpc.Response<UpdateUserRoleResponse>> {
    return await rpc.call<UpdateUserRoleResponse>('UpdateUserRole', JSON.stringify(data));
}

export async function SetUserRole(data: SetUserRoleRequest): Promise<rpc.Response<SetUserRoleResponse>> {
    return await rpc.call<SetUserRoleResponse>('SetUserRole', JSON.stringify(data));
}

export async function ListUsers(data: ListUsersRequest): Promise<rpc.Response<ListUsersResponse>> {
    return await rpc.call<ListUsersResponse>('ListUsers', JSON.stringify(data));
}

export async function CreateStudio(data: CreateStudioRequest): Promise<rpc.Response<CreateStudioResponse>> {
    return await rpc.call<CreateStudioResponse>('CreateStudio', JSON.stringify(data));
}

export async function ListMyStudios(data: ListMyStudiosRequest): Promise<rpc.Response<ListMyStudiosResponse>> {
    return await rpc.call<ListMyStudiosResponse>('ListMyStudios', JSON.stringify(data));
}

export async function ListAllStudios(data: ListAllStudiosRequest): Promise<rpc.Response<ListAllStudiosResponse>> {
    return await rpc.call<ListAllStudiosResponse>('ListAllStudios', JSON.stringify(data));
}

export async function GetStudio(data: GetStudioRequest): Promise<rpc.Response<GetStudioResponse>> {
    return await rpc.call<GetStudioResponse>('GetStudio', JSON.stringify(data));
}

export async function GetStudioDashboard(data: GetStudioDashboardRequest): Promise<rpc.Response<GetStudioDashboardResponse>> {
    return await rpc.call<GetStudioDashboardResponse>('GetStudioDashboard', JSON.stringify(data));
}

export async function UpdateStudio(data: UpdateStudioRequest): Promise<rpc.Response<UpdateStudioResponse>> {
    return await rpc.call<UpdateStudioResponse>('UpdateStudio', JSON.stringify(data));
}

export async function DeleteStudio(data: DeleteStudioRequest): Promise<rpc.Response<DeleteStudioResponse>> {
    return await rpc.call<DeleteStudioResponse>('DeleteStudio', JSON.stringify(data));
}

export async function CreateRoom(data: CreateRoomRequest): Promise<rpc.Response<CreateRoomResponse>> {
    return await rpc.call<CreateRoomResponse>('CreateRoom', JSON.stringify(data));
}

export async function ListRooms(data: ListRoomsRequest): Promise<rpc.Response<ListRoomsResponse>> {
    return await rpc.call<ListRoomsResponse>('ListRooms', JSON.stringify(data));
}

export async function GetRoomDetails(data: GetRoomDetailsRequest): Promise<rpc.Response<GetRoomDetailsResponse>> {
    return await rpc.call<GetRoomDetailsResponse>('GetRoomDetails', JSON.stringify(data));
}

export async function GetStudioRoomsForCodeSession(data: GetStudioRoomsForCodeSessionRequest): Promise<rpc.Response<GetStudioRoomsForCodeSessionResponse>> {
    return await rpc.call<GetStudioRoomsForCodeSessionResponse>('GetStudioRoomsForCodeSession', JSON.stringify(data));
}

export async function ListMyAccessibleRooms(data: ListMyAccessibleRoomsRequest): Promise<rpc.Response<ListMyAccessibleRoomsResponse>> {
    return await rpc.call<ListMyAccessibleRoomsResponse>('ListMyAccessibleRooms', JSON.stringify(data));
}

export async function GetRoomStreamKey(data: GetRoomStreamKeyRequest): Promise<rpc.Response<GetRoomStreamKeyResponse>> {
    return await rpc.call<GetRoomStreamKeyResponse>('GetRoomStreamKey', JSON.stringify(data));
}

export async function UpdateRoom(data: UpdateRoomRequest): Promise<rpc.Response<UpdateRoomResponse>> {
    return await rpc.call<UpdateRoomResponse>('UpdateRoom', JSON.stringify(data));
}

export async function RegenerateStreamKey(data: RegenerateStreamKeyRequest): Promise<rpc.Response<RegenerateStreamKeyResponse>> {
    return await rpc.call<RegenerateStreamKeyResponse>('RegenerateStreamKey', JSON.stringify(data));
}

export async function DeleteRoom(data: DeleteRoomRequest): Promise<rpc.Response<DeleteRoomResponse>> {
    return await rpc.call<DeleteRoomResponse>('DeleteRoom', JSON.stringify(data));
}

export async function AddStudioMember(data: AddStudioMemberRequest): Promise<rpc.Response<AddStudioMemberResponse>> {
    return await rpc.call<AddStudioMemberResponse>('AddStudioMember', JSON.stringify(data));
}

export async function RemoveStudioMember(data: RemoveStudioMemberRequest): Promise<rpc.Response<RemoveStudioMemberResponse>> {
    return await rpc.call<RemoveStudioMemberResponse>('RemoveStudioMember', JSON.stringify(data));
}

export async function UpdateStudioMemberRole(data: UpdateStudioMemberRoleRequest): Promise<rpc.Response<UpdateStudioMemberRoleResponse>> {
    return await rpc.call<UpdateStudioMemberRoleResponse>('UpdateStudioMemberRole', JSON.stringify(data));
}

export async function ListStudioMembersAPI(data: ListStudioMembersRequest): Promise<rpc.Response<ListStudioMembersResponse>> {
    return await rpc.call<ListStudioMembersResponse>('ListStudioMembersAPI', JSON.stringify(data));
}

export async function LeaveStudio(data: LeaveStudioRequest): Promise<rpc.Response<LeaveStudioResponse>> {
    return await rpc.call<LeaveStudioResponse>('LeaveStudio', JSON.stringify(data));
}

export async function GenerateAccessCode(data: GenerateAccessCodeRequest): Promise<rpc.Response<GenerateAccessCodeResponse>> {
    return await rpc.call<GenerateAccessCodeResponse>('GenerateAccessCode', JSON.stringify(data));
}

export async function GetCodeStreamAccess(data: GetCodeStreamAccessRequest): Promise<rpc.Response<GetCodeStreamAccessResponse>> {
    return await rpc.call<GetCodeStreamAccessResponse>('GetCodeStreamAccess', JSON.stringify(data));
}

export async function RevokeAccessCode(data: RevokeAccessCodeRequest): Promise<rpc.Response<RevokeAccessCodeResponse>> {
    return await rpc.call<RevokeAccessCodeResponse>('RevokeAccessCode', JSON.stringify(data));
}

export async function ListAccessCodes(data: ListAccessCodesRequest): Promise<rpc.Response<ListAccessCodesResponse>> {
    return await rpc.call<ListAccessCodesResponse>('ListAccessCodes', JSON.stringify(data));
}

export async function GetCodeAnalytics(data: GetCodeAnalyticsRequest): Promise<rpc.Response<GetCodeAnalyticsResponse>> {
    return await rpc.call<GetCodeAnalyticsResponse>('GetCodeAnalytics', JSON.stringify(data));
}

export async function SetCameraConfig(data: SetCameraConfigRequest): Promise<rpc.Response<SetCameraConfigResponse>> {
    return await rpc.call<SetCameraConfigResponse>('SetCameraConfig', JSON.stringify(data));
}

export async function GetCameraConfigProc(data: GetCameraConfigRequest): Promise<rpc.Response<GetCameraConfigResponse>> {
    return await rpc.call<GetCameraConfigResponse>('GetCameraConfigProc', JSON.stringify(data));
}

export async function DeleteCameraConfigProc(data: DeleteCameraConfigRequest): Promise<rpc.Response<DeleteCameraConfigResponse>> {
    return await rpc.call<DeleteCameraConfigResponse>('DeleteCameraConfigProc', JSON.stringify(data));
}

export async function GetRoomAnalytics(data: GetRoomAnalyticsRequest): Promise<rpc.Response<GetRoomAnalyticsResponse>> {
    return await rpc.call<GetRoomAnalyticsResponse>('GetRoomAnalytics', JSON.stringify(data));
}

export async function GetStudioAnalytics(data: GetStudioAnalyticsRequest): Promise<rpc.Response<GetStudioAnalyticsResponse>> {
    return await rpc.call<GetStudioAnalyticsResponse>('GetStudioAnalytics', JSON.stringify(data));
}

export async function RecalculateViewerCounts(data: RecalculateViewerCountsRequest): Promise<rpc.Response<RecalculateViewerCountsResponse>> {
    return await rpc.call<RecalculateViewerCountsResponse>('RecalculateViewerCounts', JSON.stringify(data));
}

export async function SendEmote(data: SendEmoteRequest): Promise<rpc.Response<SendEmoteResponse>> {
    return await rpc.call<SendEmoteResponse>('SendEmote', JSON.stringify(data));
}

export async function ValidateStreamKey(data: SRSAuthCallback): Promise<rpc.Response<SRSAuthResponse>> {
    return await rpc.call<SRSAuthResponse>('ValidateStreamKey', JSON.stringify(data));
}

export async function HandleStreamUnpublish(data: SRSAuthCallback): Promise<rpc.Response<SRSAuthResponse>> {
    return await rpc.call<SRSAuthResponse>('HandleStreamUnpublish', JSON.stringify(data));
}

