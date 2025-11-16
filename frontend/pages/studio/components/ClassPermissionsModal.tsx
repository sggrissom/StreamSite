import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../../server";
import { Modal } from "../../../components/Modal";
import "./ClassPermissionsModal-styles";

type ClassPermissionsModalProps = {
  isOpen: boolean;
  onClose: () => void;
  schedule: server.ClassSchedule;
  studioId: number;
};

type ModalState = {
  permissions: server.ClassPermissionWithUser[];
  studioMembers: server.MemberWithDetails[];
  isLoadingPermissions: boolean;
  isLoadingMembers: boolean;
  error: string;
  selectedUserId: number; // For add user dropdown
  isGranting: boolean;
  isRevoking: number; // Permission ID being revoked, 0 if none
};

const useModalState = vlens.declareHook((scheduleId: number): ModalState => {
  const state: ModalState = {
    permissions: [],
    studioMembers: [],
    isLoadingPermissions: true,
    isLoadingMembers: true,
    error: "",
    selectedUserId: 0,
    isGranting: false,
    isRevoking: 0,
  };

  return state;
});

async function loadPermissions(state: ModalState, scheduleId: number) {
  state.isLoadingPermissions = true;
  state.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.ListClassPermissions({ scheduleId });

  state.isLoadingPermissions = false;

  if (err || !resp) {
    state.error = err || "Failed to load permissions";
    vlens.scheduleRedraw();
    return;
  }

  state.permissions = resp.permissions || [];
  vlens.scheduleRedraw();
}

async function loadStudioMembers(state: ModalState, studioId: number) {
  state.isLoadingMembers = true;
  vlens.scheduleRedraw();

  const [resp, err] = await server.ListStudioMembersAPI({ studioId });

  state.isLoadingMembers = false;

  if (err || !resp) {
    // Don't show error for members load failure, just log it
    console.error("Failed to load studio members:", err);
    vlens.scheduleRedraw();
    return;
  }

  state.studioMembers = resp.members || [];
  vlens.scheduleRedraw();
}

function handleUserSelect(state: ModalState, e: Event) {
  state.selectedUserId = parseInt((e.target as HTMLSelectElement).value, 10);
  vlens.scheduleRedraw();
}

async function handleGrantPermission(
  state: ModalState,
  scheduleId: number,
  userId: number,
) {
  if (userId === 0) {
    state.error = "Please select a user";
    vlens.scheduleRedraw();
    return;
  }

  // Check if user already has permission
  const existing = state.permissions.find(
    (p) => p.permission.userId === userId,
  );
  if (existing) {
    state.error = "User already has permission to this class";
    vlens.scheduleRedraw();
    return;
  }

  state.isGranting = true;
  state.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.GrantClassPermission({
    scheduleId,
    userId,
    role: server.StudioRoleViewer, // Always grant viewer access
  });

  state.isGranting = false;

  if (err || !resp) {
    state.error = err || "Failed to grant permission";
    vlens.scheduleRedraw();
    return;
  }

  // Reset selection and reload permissions
  state.selectedUserId = 0;
  await loadPermissions(state, scheduleId);
}

async function handleRevokePermission(
  state: ModalState,
  scheduleId: number,
  permissionId: number,
) {
  if (
    !confirm(
      "Are you sure you want to revoke this user's access to this class?",
    )
  ) {
    return;
  }

  state.isRevoking = permissionId;
  state.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.RevokeClassPermission({ permissionId });

  state.isRevoking = 0;

  if (err || !resp || !resp.success) {
    state.error = err || "Failed to revoke permission";
    vlens.scheduleRedraw();
    return;
  }

  // Reload permissions
  await loadPermissions(state, scheduleId);
}

export function ClassPermissionsModal(props: ClassPermissionsModalProps) {
  const state = useModalState(props.schedule.id);

  // Load data on first render
  if (state.isLoadingPermissions && state.permissions.length === 0) {
    loadPermissions(state, props.schedule.id);
  }
  if (state.isLoadingMembers && state.studioMembers.length === 0) {
    loadStudioMembers(state, props.studioId);
  }

  // Get available users (not already having permission)
  const availableUsers = state.studioMembers.filter(
    (member) =>
      !state.permissions.some((p) => p.permission.userId === member.userId),
  );

  return (
    <Modal
      isOpen={props.isOpen}
      title={`Manage Permissions - ${props.schedule.name}`}
      onClose={props.onClose}
    >
      <div className="class-permissions-modal">
        {/* Add User Section */}
        <div className="permissions-add-section">
          <h3 className="permissions-section-title">Grant Access</h3>
          <div className="permissions-add-form">
            <div className="permissions-form-row">
              <div className="permissions-form-group">
                <label className="permissions-form-label">User</label>
                <select
                  className="permissions-form-select"
                  value={state.selectedUserId}
                  onChange={vlens.cachePartial(handleUserSelect, state)}
                  disabled={state.isGranting || availableUsers.length === 0}
                >
                  <option value={0}>
                    {availableUsers.length === 0
                      ? "All members have access"
                      : "Select user..."}
                  </option>
                  {availableUsers.map((member) => (
                    <option key={member.userId} value={member.userId}>
                      {member.userName} ({member.userEmail})
                    </option>
                  ))}
                </select>
              </div>

              <button
                className="btn-permissions-grant"
                onClick={() =>
                  handleGrantPermission(
                    state,
                    props.schedule.id,
                    state.selectedUserId,
                  )
                }
                disabled={
                  state.isGranting ||
                  state.selectedUserId === 0 ||
                  availableUsers.length === 0
                }
              >
                {state.isGranting ? "Granting..." : "Grant Access"}
              </button>
            </div>
          </div>
        </div>

        {/* Error Display */}
        {state.error && <div className="permissions-error">{state.error}</div>}

        {/* Current Permissions Section */}
        <div className="permissions-list-section">
          <h3 className="permissions-section-title">Current Permissions</h3>

          {state.isLoadingPermissions && (
            <div className="permissions-loading">Loading permissions...</div>
          )}

          {!state.isLoadingPermissions && state.permissions.length === 0 && (
            <div className="permissions-empty">
              No users have been granted access to this class yet.
            </div>
          )}

          {!state.isLoadingPermissions && state.permissions.length > 0 && (
            <table className="permissions-table">
              <thead>
                <tr>
                  <th>User</th>
                  <th>Email</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {state.permissions.map((perm) => (
                  <tr key={perm.permission.id}>
                    <td>{perm.userName}</td>
                    <td className="permissions-table-email">
                      {perm.userEmail}
                    </td>
                    <td>
                      <button
                        className="btn-permissions-revoke"
                        onClick={() =>
                          handleRevokePermission(
                            state,
                            props.schedule.id,
                            perm.permission.id,
                          )
                        }
                        disabled={state.isRevoking === perm.permission.id}
                      >
                        {state.isRevoking === perm.permission.id
                          ? "Revoking..."
                          : "Revoke"}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* Close Button */}
        <div className="permissions-modal-actions">
          <button className="btn-permissions-close" onClick={props.onClose}>
            Close
          </button>
        </div>
      </div>
    </Modal>
  );
}
