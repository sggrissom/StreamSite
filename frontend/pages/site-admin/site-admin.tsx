import * as preact from "preact";
import * as vlens from "vlens";
import * as core from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import { Modal } from "../../components/Modal";
import "../../styles/global";
import "./site-admin-styles";

type Data = server.ListAllStudiosResponse;

export async function fetch(route: string, prefix: string) {
  return server.ListAllStudios({});
}

// ===== Delete Studio Modal =====
type DeleteStudioModal = {
  isOpen: boolean;
  isDeleting: boolean;
  error: string;
  studioId: number;
  studioName: string;
};

const useDeleteStudioModal = vlens.declareHook(
  (): DeleteStudioModal => ({
    isOpen: false,
    isDeleting: false,
    error: "",
    studioId: 0,
    studioName: "",
  }),
);

function openDeleteStudioModal(
  modal: DeleteStudioModal,
  studioId: number,
  studioName: string,
) {
  modal.isOpen = true;
  modal.error = "";
  modal.studioId = studioId;
  modal.studioName = studioName;
  vlens.scheduleRedraw();
}

function closeDeleteStudioModal(modal: DeleteStudioModal) {
  modal.isOpen = false;
  modal.error = "";
  vlens.scheduleRedraw();
}

async function confirmDeleteStudio(modal: DeleteStudioModal) {
  modal.isDeleting = true;
  modal.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.DeleteStudio({
    studioId: modal.studioId,
  });

  modal.isDeleting = false;

  if (err) {
    modal.error = err || "Failed to delete studio";
    vlens.scheduleRedraw();
    return;
  }

  closeDeleteStudioModal(modal);
  window.location.reload();
}

// ===== Users Management =====
type UsersState = {
  users: server.UserWithStats[];
  loading: boolean;
  error: string;
  changingRoleFor: number; // userId currently being updated
};

const useUsersState = vlens.declareHook(
  (): UsersState => ({
    users: [],
    loading: false,
    error: "",
    changingRoleFor: 0,
  }),
);

async function loadUsers(state: UsersState) {
  state.loading = true;
  state.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.ListAllUsers({});

  state.loading = false;

  if (err || !resp) {
    state.error = err || "Failed to load users";
    vlens.scheduleRedraw();
    return;
  }

  state.users = resp.users || [];
  vlens.scheduleRedraw();
}

async function changeUserRole(
  state: UsersState,
  userId: number,
  newRole: number,
  myUserId: number,
) {
  if (userId === myUserId) {
    return; // Cannot change own role
  }

  state.changingRoleFor = userId;
  state.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.UpdateUserRole({
    userId: userId,
    newRole: newRole,
  });

  state.changingRoleFor = 0;

  if (err) {
    state.error = err || "Failed to update user role";
    vlens.scheduleRedraw();
    return;
  }

  // Reload users to get fresh data
  await loadUsers(state);
}

function getRoleName(role: number): string {
  switch (role) {
    case 0:
      return "User";
    case 1:
      return "Stream Admin";
    case 2:
      return "Site Admin";
    default:
      return "Unknown";
  }
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  const deleteStudioModal = useDeleteStudioModal();
  const usersState = useUsersState();

  // Check permissions
  if (!data || !data.studios) {
    return (
      <div>
        <Header />
        <main className="site-admin-container">
          <div className="site-admin-content">
            <div className="error-state">
              <div className="error-icon">⚠️</div>
              <h2>Access Denied</h2>
              <p>
                You don't have permission to access this page. Only site
                administrators can view this area.
              </p>
              <a href="/dashboard" className="btn btn-primary">
                Back to Dashboard
              </a>
            </div>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  const studios = data.studios || [];

  // Load users on first render
  if (usersState.users.length === 0 && !usersState.loading) {
    loadUsers(usersState);
  }

  return (
    <div>
      <Header />
      <main className="site-admin-container">
        <div className="site-admin-content">
          <div className="site-admin-header">
            <h1 className="site-admin-title">Site Administration</h1>
            <p className="site-admin-description">
              Manage all studios, users, and system-wide settings
            </p>
          </div>

          {/* Studios Management Section */}
          <div className="admin-section">
            <div className="section-header">
              <h2 className="section-title">Studios Management</h2>
              <a href="/studios" className="btn btn-primary btn-sm">
                Create Studio
              </a>
            </div>

            {studios.length === 0 ? (
              <div className="empty-state">
                <p>No studios have been created yet.</p>
              </div>
            ) : (
              <div className="studios-table-container">
                <table className="admin-table">
                  <thead>
                    <tr>
                      <th>Studio Name</th>
                      <th>Owner</th>
                      <th>Created</th>
                      <th>Rooms</th>
                      <th>Members</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {studios.map((studio) => (
                      <tr key={studio.id}>
                        <td>
                          <div className="studio-name-cell">
                            <strong>{studio.name}</strong>
                            {studio.description && (
                              <div className="studio-description-preview">
                                {studio.description}
                              </div>
                            )}
                          </div>
                        </td>
                        <td>
                          <div className="owner-cell">
                            <div>{studio.ownerName}</div>
                            <div className="email-text">
                              {studio.ownerEmail}
                            </div>
                          </div>
                        </td>
                        <td>
                          {new Date(studio.creation).toLocaleDateString()}
                        </td>
                        <td>{studio.roomCount}</td>
                        <td>{studio.memberCount}</td>
                        <td className="actions-cell">
                          <a
                            href={`/studio/${studio.id}`}
                            className="btn btn-secondary btn-sm"
                          >
                            View Studio
                          </a>
                          <button
                            className="btn btn-danger btn-sm"
                            onClick={() =>
                              openDeleteStudioModal(
                                deleteStudioModal,
                                studio.id,
                                studio.name,
                              )
                            }
                          >
                            Delete
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          {/* Users Management Section */}
          <div className="admin-section">
            <div className="section-header">
              <h2 className="section-title">Users Management</h2>
            </div>

            {usersState.error && (
              <div className="error-message">{usersState.error}</div>
            )}

            {usersState.loading ? (
              <div className="loading-state">Loading users...</div>
            ) : usersState.users.length === 0 ? (
              <div className="empty-state">
                <p>No users found.</p>
              </div>
            ) : (
              <div className="users-table-container">
                <table className="admin-table">
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Email</th>
                      <th>Role</th>
                      <th>Studios</th>
                      <th>Created</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {usersState.users.map((user) => (
                      <tr key={user.id}>
                        <td>{user.name}</td>
                        <td>{user.email}</td>
                        <td>
                          <span className={`role-badge role-${user.role}`}>
                            {getRoleName(user.role)}
                          </span>
                        </td>
                        <td>{user.studioCount}</td>
                        <td>{new Date(user.creation).toLocaleDateString()}</td>
                        <td className="actions-cell">
                          {/* Get current user ID from first user (yourself) - you're always the site admin loading this */}
                          {user.id === 1 ? (
                            <span className="text-muted">(You)</span>
                          ) : (
                            <select
                              className="role-select"
                              disabled={usersState.changingRoleFor === user.id}
                              onChange={(e) => {
                                const target = e.target as HTMLSelectElement;
                                changeUserRole(
                                  usersState,
                                  user.id,
                                  parseInt(target.value),
                                  1, // Site admin is always user ID 1
                                );
                              }}
                              value={user.role}
                            >
                              <option value={0}>User</option>
                              <option value={1}>Stream Admin</option>
                              <option value={2}>Site Admin</option>
                            </select>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      </main>

      {/* Delete Studio Modal */}
      <Modal
        isOpen={deleteStudioModal.isOpen}
        title="Delete Studio"
        onClose={() => closeDeleteStudioModal(deleteStudioModal)}
        error={deleteStudioModal.error}
        footer={
          <>
            <button
              className="btn btn-secondary"
              onClick={() => closeDeleteStudioModal(deleteStudioModal)}
              disabled={deleteStudioModal.isDeleting}
            >
              Cancel
            </button>
            <button
              className="btn btn-danger"
              onClick={() => confirmDeleteStudio(deleteStudioModal)}
              disabled={deleteStudioModal.isDeleting}
            >
              {deleteStudioModal.isDeleting ? "Deleting..." : "Delete Studio"}
            </button>
          </>
        }
      >
        <div className="delete-confirmation">
          <p className="confirmation-text">
            ⚠️ Are you sure you want to delete this studio? This will also
            delete all rooms, streams, and memberships. This action cannot be
            undone.
          </p>
          <div className="studio-info">
            <strong>Studio:</strong> {deleteStudioModal.studioName}
          </div>
        </div>
      </Modal>

      <Footer />
    </div>
  );
}
