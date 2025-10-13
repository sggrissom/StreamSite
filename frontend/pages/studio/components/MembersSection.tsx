import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../../server";
import { Modal } from "../../../components/Modal";

type Studio = {
  id: number;
  name: string;
};

type Member = {
  userId: number;
  userName: string;
  userEmail: string;
  role: number;
  roleName: string;
  joinedAt: string;
};

type MembersSectionProps = {
  studio: Studio;
  members: Member[];
  canManageRooms: boolean;
};

function getRoleBadgeClass(role: number): string {
  return `studio-role role-${role}`;
}

// ===== Add Member Modal =====
type AddMemberModal = {
  isOpen: boolean;
  isSubmitting: boolean;
  error: string;
  studioId: number;
  userEmail: string;
  role: number;
};

const useAddMemberModal = vlens.declareHook(
  (): AddMemberModal => ({
    isOpen: false,
    isSubmitting: false,
    error: "",
    studioId: 0,
    userEmail: "",
    role: 1,
  }),
);

function openAddMemberModal(modal: AddMemberModal, studioId: number) {
  modal.isOpen = true;
  modal.error = "";
  modal.studioId = studioId;
  modal.userEmail = "";
  modal.role = 1;
  vlens.scheduleRedraw();
}

function closeAddMemberModal(modal: AddMemberModal) {
  modal.isOpen = false;
  modal.error = "";
  vlens.scheduleRedraw();
}

async function submitAddMember(modal: AddMemberModal) {
  if (!modal.userEmail.trim()) {
    modal.error = "Email is required";
    vlens.scheduleRedraw();
    return;
  }

  modal.isSubmitting = true;
  modal.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.AddStudioMember({
    studioId: modal.studioId,
    userId: 0, // Optional: using userEmail instead
    userEmail: modal.userEmail.trim(),
    role: modal.role,
  });

  modal.isSubmitting = false;

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to add member";
    vlens.scheduleRedraw();
    return;
  }

  closeAddMemberModal(modal);
  window.location.reload();
}

// ===== Change Role Modal =====
type ChangeRoleModal = {
  isOpen: boolean;
  isSubmitting: boolean;
  error: string;
  studioId: number;
  userId: number;
  userName: string;
  currentRole: number;
  newRole: number;
};

const useChangeRoleModal = vlens.declareHook(
  (): ChangeRoleModal => ({
    isOpen: false,
    isSubmitting: false,
    error: "",
    studioId: 0,
    userId: 0,
    userName: "",
    currentRole: 0,
    newRole: 0,
  }),
);

function openChangeRoleModal(
  modal: ChangeRoleModal,
  studioId: number,
  userId: number,
  userName: string,
  currentRole: number,
) {
  modal.isOpen = true;
  modal.error = "";
  modal.studioId = studioId;
  modal.userId = userId;
  modal.userName = userName;
  modal.currentRole = currentRole;
  modal.newRole = currentRole;
  vlens.scheduleRedraw();
}

function closeChangeRoleModal(modal: ChangeRoleModal) {
  modal.isOpen = false;
  modal.error = "";
  vlens.scheduleRedraw();
}

async function submitChangeRole(modal: ChangeRoleModal) {
  if (modal.newRole === modal.currentRole) {
    modal.error = "Please select a different role";
    vlens.scheduleRedraw();
    return;
  }

  modal.isSubmitting = true;
  modal.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.UpdateStudioMemberRole({
    studioId: modal.studioId,
    userId: modal.userId,
    newRole: modal.newRole,
  });

  modal.isSubmitting = false;

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to update role";
    vlens.scheduleRedraw();
    return;
  }

  closeChangeRoleModal(modal);
  window.location.reload();
}

// ===== Remove Member Modal =====
type RemoveMemberModal = {
  isOpen: boolean;
  isRemoving: boolean;
  error: string;
  studioId: number;
  userId: number;
  userName: string;
};

const useRemoveMemberModal = vlens.declareHook(
  (): RemoveMemberModal => ({
    isOpen: false,
    isRemoving: false,
    error: "",
    studioId: 0,
    userId: 0,
    userName: "",
  }),
);

function openRemoveMemberModal(
  modal: RemoveMemberModal,
  studioId: number,
  userId: number,
  userName: string,
) {
  modal.isOpen = true;
  modal.error = "";
  modal.studioId = studioId;
  modal.userId = userId;
  modal.userName = userName;
  vlens.scheduleRedraw();
}

function closeRemoveMemberModal(modal: RemoveMemberModal) {
  modal.isOpen = false;
  modal.error = "";
  vlens.scheduleRedraw();
}

async function confirmRemoveMember(modal: RemoveMemberModal) {
  modal.isRemoving = true;
  modal.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.RemoveStudioMember({
    studioId: modal.studioId,
    userId: modal.userId,
  });

  modal.isRemoving = false;

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to remove member";
    vlens.scheduleRedraw();
    return;
  }

  closeRemoveMemberModal(modal);
  window.location.reload();
}

// ===== Component =====
export function MembersSection(
  props: MembersSectionProps,
): preact.ComponentChild {
  const { studio, members, canManageRooms } = props;
  const addMemberModal = useAddMemberModal();
  const changeRoleModal = useChangeRoleModal();
  const removeMemberModal = useRemoveMemberModal();

  return (
    <>
      <div className="members-section">
        <div className="members-header">
          <h2 className="section-title">Members</h2>
          {canManageRooms && (
            <button
              className="btn btn-primary btn-sm"
              onClick={() => openAddMemberModal(addMemberModal, studio.id)}
            >
              Add Member
            </button>
          )}
        </div>

        {members.length === 0 ? (
          <div className="members-empty">
            <p>No members in this studio.</p>
          </div>
        ) : (
          <div className="members-table-container">
            <table className="members-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Email</th>
                  <th>Role</th>
                  <th>Joined</th>
                  {canManageRooms && <th>Actions</th>}
                </tr>
              </thead>
              <tbody>
                {members.map((member: any) => (
                  <tr key={member.userId}>
                    <td>{member.userName}</td>
                    <td>{member.userEmail}</td>
                    <td>
                      <span className={getRoleBadgeClass(member.role)}>
                        {member.roleName}
                      </span>
                    </td>
                    <td>{new Date(member.joinedAt).toLocaleDateString()}</td>
                    {canManageRooms && (
                      <td className="members-actions">
                        {member.role !== 3 && (
                          <>
                            <button
                              className="btn btn-secondary btn-sm"
                              onClick={() =>
                                openChangeRoleModal(
                                  changeRoleModal,
                                  studio.id,
                                  member.userId,
                                  member.userName,
                                  member.role,
                                )
                              }
                            >
                              Change Role
                            </button>
                            <button
                              className="btn btn-danger btn-sm"
                              onClick={() =>
                                openRemoveMemberModal(
                                  removeMemberModal,
                                  studio.id,
                                  member.userId,
                                  member.userName,
                                )
                              }
                            >
                              Remove
                            </button>
                          </>
                        )}
                        {member.role === 3 && (
                          <span className="member-owner-label">Owner</span>
                        )}
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Modals */}
      <Modal
        isOpen={addMemberModal.isOpen}
        title="Add Studio Member"
        onClose={() => closeAddMemberModal(addMemberModal)}
        error={addMemberModal.error}
        footer={
          <>
            <button
              className="btn btn-secondary"
              onClick={() => closeAddMemberModal(addMemberModal)}
              disabled={addMemberModal.isSubmitting}
            >
              Cancel
            </button>
            <button
              className="btn btn-primary"
              onClick={() => submitAddMember(addMemberModal)}
              disabled={addMemberModal.isSubmitting}
            >
              {addMemberModal.isSubmitting ? "Adding..." : "Add Member"}
            </button>
          </>
        }
      >
        <div className="form-group">
          <label htmlFor="member-email">User Email *</label>
          <input
            id="member-email"
            type="email"
            className="form-input"
            placeholder="user@example.com"
            {...vlens.attrsBindInput(vlens.ref(addMemberModal, "userEmail"))}
            disabled={addMemberModal.isSubmitting}
          />
          <small className="form-help">
            Enter the email address of the user to add
          </small>
        </div>

        <div className="form-group">
          <label htmlFor="member-role">Role *</label>
          <select
            id="member-role"
            className="form-input"
            {...vlens.attrsBindInput(vlens.ref(addMemberModal, "role"))}
            disabled={addMemberModal.isSubmitting}
          >
            <option value={0}>Viewer - Can watch streams</option>
            <option value={1}>Member - Can stream</option>
            <option value={2}>Admin - Can manage rooms and members</option>
          </select>
        </div>
      </Modal>

      <Modal
        isOpen={changeRoleModal.isOpen}
        title="Change Member Role"
        onClose={() => closeChangeRoleModal(changeRoleModal)}
        error={changeRoleModal.error}
        footer={
          <>
            <button
              className="btn btn-secondary"
              onClick={() => closeChangeRoleModal(changeRoleModal)}
              disabled={changeRoleModal.isSubmitting}
            >
              Cancel
            </button>
            <button
              className="btn btn-primary"
              onClick={() => submitChangeRole(changeRoleModal)}
              disabled={changeRoleModal.isSubmitting}
            >
              {changeRoleModal.isSubmitting ? "Updating..." : "Update Role"}
            </button>
          </>
        }
      >
        <div className="form-group">
          <label>Member</label>
          <div className="member-info">{changeRoleModal.userName}</div>
        </div>

        <div className="form-group">
          <label htmlFor="change-role">New Role *</label>
          <select
            id="change-role"
            className="form-input"
            {...vlens.attrsBindInput(vlens.ref(changeRoleModal, "newRole"))}
            disabled={changeRoleModal.isSubmitting}
          >
            <option value={0}>Viewer - Can watch streams</option>
            <option value={1}>Member - Can stream</option>
            <option value={2}>Admin - Can manage rooms and members</option>
          </select>
        </div>
      </Modal>

      <Modal
        isOpen={removeMemberModal.isOpen}
        title="Remove Studio Member"
        onClose={() => closeRemoveMemberModal(removeMemberModal)}
        error={removeMemberModal.error}
        footer={
          <>
            <button
              className="btn btn-secondary"
              onClick={() => closeRemoveMemberModal(removeMemberModal)}
              disabled={removeMemberModal.isRemoving}
            >
              Cancel
            </button>
            <button
              className="btn btn-danger"
              onClick={() => confirmRemoveMember(removeMemberModal)}
              disabled={removeMemberModal.isRemoving}
            >
              {removeMemberModal.isRemoving ? "Removing..." : "Remove Member"}
            </button>
          </>
        }
      >
        <div className="remove-confirmation">
          <p className="confirmation-text">
            ⚠️ Are you sure you want to remove this member from the studio?
          </p>
          <div className="member-info">
            <strong>Member:</strong> {removeMemberModal.userName}
          </div>
        </div>
      </Modal>
    </>
  );
}
