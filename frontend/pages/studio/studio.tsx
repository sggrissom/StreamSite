import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import { Modal } from "../../components/Modal";
import "../../styles/global";
import "./studio-styles";

type Data = server.GetStudioDashboardResponse;

export async function fetch(route: string, prefix: string) {
  // Extract studio ID from route (e.g., "/studio/123" -> "123")
  const studioIdStr = route
    .replace(prefix, "")
    .replace(/^\//, "")
    .split("/")[0];
  const studioId = parseInt(studioIdStr, 10);

  return server.GetStudioDashboard({ studioId });
}

function getRoleBadgeClass(role: number): string {
  return `studio-role role-${role}`;
}

type CreateRoomModal = {
  isOpen: boolean;
  isSubmitting: boolean;
  error: string;
  name: string;
  studioId: number;
};

const useCreateRoomModal = vlens.declareHook(
  (): CreateRoomModal => ({
    isOpen: false,
    isSubmitting: false,
    error: "",
    name: "",
    studioId: 0,
  }),
);

function openRoomModal(modal: CreateRoomModal, studioId: number) {
  modal.isOpen = true;
  modal.error = "";
  modal.name = "";
  modal.studioId = studioId;
  vlens.scheduleRedraw();
}

function closeRoomModal(modal: CreateRoomModal) {
  modal.isOpen = false;
  modal.error = "";
  vlens.scheduleRedraw();
}

async function submitCreateRoom(modal: CreateRoomModal) {
  if (!modal.name.trim()) {
    modal.error = "Room name is required";
    vlens.scheduleRedraw();
    return;
  }

  modal.isSubmitting = true;
  modal.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.CreateRoom({
    studioId: modal.studioId,
    name: modal.name.trim(),
  });

  modal.isSubmitting = false;

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to create room";
    vlens.scheduleRedraw();
    return;
  }

  // Success - close modal and reload page
  closeRoomModal(modal);
  window.location.reload();
}

type StreamKeyModal = {
  isOpen: boolean;
  isLoading: boolean;
  error: string;
  roomId: number;
  roomName: string;
  streamKey: string;
  showConfirmRegenerate: boolean;
  isRegenerating: boolean;
  copySuccess: boolean;
};

const useStreamKeyModal = vlens.declareHook(
  (): StreamKeyModal => ({
    isOpen: false,
    isLoading: false,
    error: "",
    roomId: 0,
    roomName: "",
    streamKey: "",
    showConfirmRegenerate: false,
    isRegenerating: false,
    copySuccess: false,
  }),
);

async function openStreamKeyModal(
  modal: StreamKeyModal,
  roomId: number,
  roomName: string,
) {
  modal.isOpen = true;
  modal.isLoading = true;
  modal.error = "";
  modal.roomId = roomId;
  modal.roomName = roomName;
  modal.streamKey = "";
  modal.showConfirmRegenerate = false;
  modal.isRegenerating = false;
  modal.copySuccess = false;
  vlens.scheduleRedraw();

  // Fetch stream key
  const [resp, err] = await server.GetRoomStreamKey({ roomId });

  modal.isLoading = false;

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to load stream key";
    vlens.scheduleRedraw();
    return;
  }

  modal.streamKey = resp.streamKey || "";
  vlens.scheduleRedraw();
}

function closeStreamKeyModal(modal: StreamKeyModal) {
  modal.isOpen = false;
  modal.error = "";
  modal.streamKey = "";
  modal.showConfirmRegenerate = false;
  vlens.scheduleRedraw();
}

async function copyStreamKey(modal: StreamKeyModal) {
  try {
    await navigator.clipboard.writeText(modal.streamKey);
    modal.copySuccess = true;
    vlens.scheduleRedraw();

    // Reset copy success after 2 seconds
    setTimeout(() => {
      modal.copySuccess = false;
      vlens.scheduleRedraw();
    }, 2000);
  } catch (err) {
    modal.error = "Failed to copy to clipboard";
    vlens.scheduleRedraw();
  }
}

function showRegenerateConfirmation(modal: StreamKeyModal) {
  modal.showConfirmRegenerate = true;
  modal.error = "";
  vlens.scheduleRedraw();
}

function hideRegenerateConfirmation(modal: StreamKeyModal) {
  modal.showConfirmRegenerate = false;
  vlens.scheduleRedraw();
}

async function confirmRegenerateStreamKey(modal: StreamKeyModal) {
  modal.isRegenerating = true;
  modal.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.RegenerateStreamKey({
    roomId: modal.roomId,
  });

  modal.isRegenerating = false;
  modal.showConfirmRegenerate = false;

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to regenerate stream key";
    vlens.scheduleRedraw();
    return;
  }

  modal.streamKey = resp.streamKey || "";
  modal.copySuccess = false;
  vlens.scheduleRedraw();
}

type EditRoomModal = {
  isOpen: boolean;
  isSubmitting: boolean;
  error: string;
  roomId: number;
  roomName: string;
};

const useEditRoomModal = vlens.declareHook(
  (): EditRoomModal => ({
    isOpen: false,
    isSubmitting: false,
    error: "",
    roomId: 0,
    roomName: "",
  }),
);

function openEditRoomModal(
  modal: EditRoomModal,
  roomId: number,
  currentName: string,
) {
  modal.isOpen = true;
  modal.error = "";
  modal.roomId = roomId;
  modal.roomName = currentName;
  vlens.scheduleRedraw();
}

function closeEditRoomModal(modal: EditRoomModal) {
  modal.isOpen = false;
  modal.error = "";
  vlens.scheduleRedraw();
}

async function submitEditRoom(modal: EditRoomModal) {
  if (!modal.roomName.trim()) {
    modal.error = "Room name is required";
    vlens.scheduleRedraw();
    return;
  }

  modal.isSubmitting = true;
  modal.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.UpdateRoom({
    roomId: modal.roomId,
    name: modal.roomName.trim(),
  });

  modal.isSubmitting = false;

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to update room";
    vlens.scheduleRedraw();
    return;
  }

  // Success - close modal and reload page
  closeEditRoomModal(modal);
  window.location.reload();
}

type DeleteRoomModal = {
  isOpen: boolean;
  isDeleting: boolean;
  error: string;
  roomId: number;
  roomName: string;
  isActive: boolean;
};

const useDeleteRoomModal = vlens.declareHook(
  (): DeleteRoomModal => ({
    isOpen: false,
    isDeleting: false,
    error: "",
    roomId: 0,
    roomName: "",
    isActive: false,
  }),
);

function openDeleteRoomModal(
  modal: DeleteRoomModal,
  roomId: number,
  roomName: string,
  isActive: boolean,
) {
  modal.isOpen = true;
  modal.error = "";
  modal.roomId = roomId;
  modal.roomName = roomName;
  modal.isActive = isActive;
  vlens.scheduleRedraw();
}

function closeDeleteRoomModal(modal: DeleteRoomModal) {
  modal.isOpen = false;
  modal.error = "";
  vlens.scheduleRedraw();
}

async function confirmDeleteRoom(modal: DeleteRoomModal) {
  modal.isDeleting = true;
  modal.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.DeleteRoom({
    roomId: modal.roomId,
  });

  modal.isDeleting = false;

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to delete room";
    vlens.scheduleRedraw();
    return;
  }

  // Success - close modal and reload page
  closeDeleteRoomModal(modal);
  window.location.reload();
}

type EditStudioModal = {
  isOpen: boolean;
  isSubmitting: boolean;
  error: string;
  studioId: number;
  studioName: string;
  studioDescription: string;
  maxRooms: number;
};

const useEditStudioModal = vlens.declareHook(
  (): EditStudioModal => ({
    isOpen: false,
    isSubmitting: false,
    error: "",
    studioId: 0,
    studioName: "",
    studioDescription: "",
    maxRooms: 5,
  }),
);

function openEditStudioModal(
  modal: EditStudioModal,
  studioId: number,
  name: string,
  description: string,
  maxRooms: number,
) {
  modal.isOpen = true;
  modal.error = "";
  modal.studioId = studioId;
  modal.studioName = name;
  modal.studioDescription = description;
  modal.maxRooms = maxRooms;
  vlens.scheduleRedraw();
}

function closeEditStudioModal(modal: EditStudioModal) {
  modal.isOpen = false;
  modal.error = "";
  vlens.scheduleRedraw();
}

async function submitEditStudio(modal: EditStudioModal) {
  if (!modal.studioName.trim()) {
    modal.error = "Studio name is required";
    vlens.scheduleRedraw();
    return;
  }

  if (modal.maxRooms < 1) {
    modal.error = "Max rooms must be at least 1";
    vlens.scheduleRedraw();
    return;
  }

  modal.isSubmitting = true;
  modal.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.UpdateStudio({
    studioId: modal.studioId,
    name: modal.studioName.trim(),
    description: modal.studioDescription.trim(),
    maxRooms: modal.maxRooms,
  });

  modal.isSubmitting = false;

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to update studio";
    vlens.scheduleRedraw();
    return;
  }

  // Success - close modal and reload page
  closeEditStudioModal(modal);
  window.location.reload();
}

type DeleteStudioModal = {
  isOpen: boolean;
  isDeleting: boolean;
  error: string;
  studioId: number;
  studioName: string;
  isOwner: boolean;
};

const useDeleteStudioModal = vlens.declareHook(
  (): DeleteStudioModal => ({
    isOpen: false,
    isDeleting: false,
    error: "",
    studioId: 0,
    studioName: "",
    isOwner: false,
  }),
);

function openDeleteStudioModal(
  modal: DeleteStudioModal,
  studioId: number,
  studioName: string,
  isOwner: boolean,
) {
  modal.isOpen = true;
  modal.error = "";
  modal.studioId = studioId;
  modal.studioName = studioName;
  modal.isOwner = isOwner;
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

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to delete studio";
    vlens.scheduleRedraw();
    return;
  }

  // Success - redirect to studios list
  closeDeleteStudioModal(modal);
  window.location.href = "/studios";
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  const roomModal = useCreateRoomModal();
  const streamKeyModal = useStreamKeyModal();
  const editRoomModal = useEditRoomModal();
  const deleteRoomModal = useDeleteRoomModal();
  const editStudioModal = useEditStudioModal();
  const deleteStudioModal = useDeleteStudioModal();

  // Handle errors or missing data
  if (!data || !data.success) {
    return (
      <div>
        <Header />
        <main className="studio-container">
          <div className="studio-content">
            <div className="error-state">
              <div className="error-icon">⚠️</div>
              <h2>Studio Not Found</h2>
              <p>
                {data?.error ||
                  "The studio you're looking for doesn't exist or you don't have permission to view it."}
              </p>
              <a href="/studios" className="btn btn-primary">
                Back to Studios
              </a>
            </div>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  const studio = data.studio;
  const rooms = data.rooms || [];
  const myRole = data.myRole;
  const myRoleName = data.myRoleName;

  // Check if user can manage rooms (Admin or Owner)
  const canManageRooms = myRole >= 2; // Admin or Owner

  return (
    <div>
      <Header />
      <main className="studio-container">
        <div className="studio-content">
          {/* Breadcrumb */}
          <div className="breadcrumb">
            <a href="/studios">Studios</a>
            <span className="breadcrumb-separator">/</span>
            <span className="breadcrumb-current">{studio.name}</span>
          </div>

          {/* Studio Header */}
          <div className="studio-header">
            <div className="studio-header-main">
              <h1 className="studio-title">{studio.name}</h1>
              <span className={getRoleBadgeClass(myRole)}>{myRoleName}</span>
            </div>
            {studio.description && (
              <p className="studio-description">{studio.description}</p>
            )}
          </div>

          {/* Studio Metadata */}
          <div className="studio-metadata">
            <div className="metadata-card">
              <div className="metadata-label">Max Rooms</div>
              <div className="metadata-value">{studio.maxRooms}</div>
            </div>
            <div className="metadata-card">
              <div className="metadata-label">Total Rooms</div>
              <div className="metadata-value">{rooms.length}</div>
            </div>
            <div className="metadata-card">
              <div className="metadata-label">Active Rooms</div>
              <div className="metadata-value">
                {rooms.filter((r) => r.isActive).length}
              </div>
            </div>
          </div>

          {/* Action Buttons */}
          {canManageRooms && (
            <div className="studio-actions">
              <button
                className="btn btn-secondary"
                onClick={() =>
                  openEditStudioModal(
                    editStudioModal,
                    studio.id,
                    studio.name,
                    studio.description || "",
                    studio.maxRooms,
                  )
                }
              >
                Edit Studio
              </button>
              {myRole === 3 && (
                <button
                  className="btn btn-danger"
                  onClick={() =>
                    openDeleteStudioModal(
                      deleteStudioModal,
                      studio.id,
                      studio.name,
                      true,
                    )
                  }
                >
                  Delete Studio
                </button>
              )}
              <button className="btn btn-secondary" disabled>
                Manage Members
              </button>
            </div>
          )}

          {/* Rooms Section */}
          <div className="rooms-section">
            <div className="rooms-header">
              <h2 className="section-title">Rooms</h2>
              {canManageRooms && rooms.length < studio.maxRooms && (
                <button
                  className="btn btn-primary btn-sm"
                  onClick={() => openRoomModal(roomModal, studio.id)}
                >
                  Create Room
                </button>
              )}
            </div>

            {rooms.length === 0 ? (
              <div className="rooms-empty">
                <div className="empty-icon">🎬</div>
                <h3>No Rooms Yet</h3>
                <p>
                  {canManageRooms
                    ? "Create your first room to start streaming."
                    : "This studio doesn't have any rooms yet."}
                </p>
                {canManageRooms && (
                  <button
                    className="btn btn-primary"
                    onClick={() => openRoomModal(roomModal, studio.id)}
                  >
                    Create First Room
                  </button>
                )}
              </div>
            ) : (
              <div className="rooms-grid">
                {rooms.map((room) => (
                  <div key={room.id} className="room-card">
                    <div className="room-header">
                      <div className="room-number">Room {room.roomNumber}</div>
                      {room.isActive && (
                        <span className="room-status active">🔴 Live</span>
                      )}
                    </div>

                    <h3 className="room-name">{room.name}</h3>

                    <div className="room-meta">
                      <span className="meta-item">
                        <span className="meta-label">Created:</span>
                        <span className="meta-value">
                          {new Date(room.creation).toLocaleDateString()}
                        </span>
                      </span>
                    </div>

                    <div className="room-actions">
                      {canManageRooms && (
                        <>
                          <button
                            className="btn btn-secondary btn-sm"
                            onClick={() =>
                              openStreamKeyModal(
                                streamKeyModal,
                                room.id,
                                room.name,
                              )
                            }
                          >
                            View Stream Key
                          </button>
                          <button
                            className="btn btn-secondary btn-sm"
                            onClick={() =>
                              openEditRoomModal(
                                editRoomModal,
                                room.id,
                                room.name,
                              )
                            }
                          >
                            Edit
                          </button>
                          <button
                            className="btn btn-danger btn-sm"
                            onClick={() =>
                              openDeleteRoomModal(
                                deleteRoomModal,
                                room.id,
                                room.name,
                                room.isActive,
                              )
                            }
                          >
                            Delete
                          </button>
                        </>
                      )}
                      {!canManageRooms && room.isActive && (
                        <button className="btn btn-primary btn-sm" disabled>
                          Watch Stream
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}

            {canManageRooms && rooms.length >= studio.maxRooms && (
              <div className="rooms-limit-notice">
                <p>
                  ⚠️ You've reached the maximum number of rooms (
                  {studio.maxRooms}
                  ). To create more rooms, increase the limit in studio
                  settings.
                </p>
              </div>
            )}
          </div>

          {/* Create Room Modal */}
          <Modal
            isOpen={roomModal.isOpen}
            title="Create New Room"
            onClose={() => closeRoomModal(roomModal)}
            error={roomModal.error}
            footer={
              <>
                <button
                  className="btn btn-secondary"
                  onClick={() => closeRoomModal(roomModal)}
                  disabled={roomModal.isSubmitting}
                >
                  Cancel
                </button>
                <button
                  className="btn btn-primary"
                  onClick={() => submitCreateRoom(roomModal)}
                  disabled={roomModal.isSubmitting}
                >
                  {roomModal.isSubmitting ? "Creating..." : "Create Room"}
                </button>
              </>
            }
          >
            <div className="form-group">
              <label htmlFor="room-name">Room Name *</label>
              <input
                id="room-name"
                type="text"
                className="form-input"
                placeholder="e.g., Main Stage, Studio A"
                {...vlens.attrsBindInput(vlens.ref(roomModal, "name"))}
                disabled={roomModal.isSubmitting}
              />
              <small className="form-help">
                Choose a descriptive name for this streaming room
              </small>
            </div>
          </Modal>

          {/* Stream Key Modal */}
          <Modal
            isOpen={streamKeyModal.isOpen}
            title="Stream Key"
            onClose={() => closeStreamKeyModal(streamKeyModal)}
            error={streamKeyModal.error}
            footer={
              <button
                className="btn btn-secondary"
                onClick={() => closeStreamKeyModal(streamKeyModal)}
              >
                Close
              </button>
            }
          >
            <div className="form-group">
              <label>Room</label>
              <div className="stream-key-room-name">
                {streamKeyModal.roomName}
              </div>
            </div>

            {streamKeyModal.isLoading ? (
              <div className="stream-key-loading">Loading...</div>
            ) : (
              streamKeyModal.streamKey && (
                <>
                  <div className="form-group">
                    <label>Stream Key</label>
                    <div className="stream-key-display">
                      {streamKeyModal.streamKey}
                    </div>
                    <small className="form-help">
                      Use this key in your streaming software (OBS, etc.)
                    </small>
                  </div>

                  <div className="stream-key-actions">
                    <button
                      className="btn btn-primary"
                      onClick={() => copyStreamKey(streamKeyModal)}
                      disabled={streamKeyModal.copySuccess}
                    >
                      {streamKeyModal.copySuccess
                        ? "Copied!"
                        : "Copy to Clipboard"}
                    </button>
                    <button
                      className="btn btn-secondary"
                      onClick={() => showRegenerateConfirmation(streamKeyModal)}
                      disabled={streamKeyModal.isRegenerating}
                    >
                      Regenerate Key
                    </button>
                  </div>

                  {streamKeyModal.showConfirmRegenerate && (
                    <div className="confirmation-dialog">
                      <p className="confirmation-text">
                        ⚠️ Are you sure you want to regenerate the stream key?
                        The old key will stop working immediately.
                      </p>
                      <div className="confirmation-actions">
                        <button
                          className="btn btn-secondary btn-sm"
                          onClick={() =>
                            hideRegenerateConfirmation(streamKeyModal)
                          }
                          disabled={streamKeyModal.isRegenerating}
                        >
                          Cancel
                        </button>
                        <button
                          className="btn btn-danger btn-sm"
                          onClick={() =>
                            confirmRegenerateStreamKey(streamKeyModal)
                          }
                          disabled={streamKeyModal.isRegenerating}
                        >
                          {streamKeyModal.isRegenerating
                            ? "Regenerating..."
                            : "Yes, Regenerate"}
                        </button>
                      </div>
                    </div>
                  )}
                </>
              )
            )}
          </Modal>

          {/* Edit Room Modal */}
          <Modal
            isOpen={editRoomModal.isOpen}
            title="Edit Room"
            onClose={() => closeEditRoomModal(editRoomModal)}
            error={editRoomModal.error}
            footer={
              <>
                <button
                  className="btn btn-secondary"
                  onClick={() => closeEditRoomModal(editRoomModal)}
                  disabled={editRoomModal.isSubmitting}
                >
                  Cancel
                </button>
                <button
                  className="btn btn-primary"
                  onClick={() => submitEditRoom(editRoomModal)}
                  disabled={editRoomModal.isSubmitting}
                >
                  {editRoomModal.isSubmitting ? "Saving..." : "Save Changes"}
                </button>
              </>
            }
          >
            <div className="form-group">
              <label htmlFor="edit-room-name">Room Name *</label>
              <input
                id="edit-room-name"
                type="text"
                className="form-input"
                placeholder="Enter room name"
                {...vlens.attrsBindInput(vlens.ref(editRoomModal, "roomName"))}
                disabled={editRoomModal.isSubmitting}
              />
            </div>
          </Modal>

          {/* Delete Room Modal */}
          <Modal
            isOpen={deleteRoomModal.isOpen}
            title="Delete Room"
            onClose={() => closeDeleteRoomModal(deleteRoomModal)}
            error={deleteRoomModal.error}
            footer={
              <>
                <button
                  className="btn btn-secondary"
                  onClick={() => closeDeleteRoomModal(deleteRoomModal)}
                  disabled={deleteRoomModal.isDeleting}
                >
                  {deleteRoomModal.isActive ? "Close" : "Cancel"}
                </button>
                {!deleteRoomModal.isActive && (
                  <button
                    className="btn btn-danger"
                    onClick={() => confirmDeleteRoom(deleteRoomModal)}
                    disabled={deleteRoomModal.isDeleting}
                  >
                    {deleteRoomModal.isDeleting ? "Deleting..." : "Delete Room"}
                  </button>
                )}
              </>
            }
          >
            {deleteRoomModal.isActive ? (
              <div className="delete-warning">
                <p className="warning-text">
                  ⚠️ Cannot delete this room while it is actively streaming.
                  Please stop the stream first.
                </p>
                <div className="room-info">
                  <strong>Room:</strong> {deleteRoomModal.roomName}
                </div>
              </div>
            ) : (
              <div className="delete-confirmation">
                <p className="confirmation-text">
                  ⚠️ Are you sure you want to delete this room? This action
                  cannot be undone.
                </p>
                <div className="room-info">
                  <strong>Room:</strong> {deleteRoomModal.roomName}
                </div>
              </div>
            )}
          </Modal>

          {/* Edit Studio Modal */}
          <Modal
            isOpen={editStudioModal.isOpen}
            title="Edit Studio"
            onClose={() => closeEditStudioModal(editStudioModal)}
            error={editStudioModal.error}
            footer={
              <>
                <button
                  className="btn btn-secondary"
                  onClick={() => closeEditStudioModal(editStudioModal)}
                  disabled={editStudioModal.isSubmitting}
                >
                  Cancel
                </button>
                <button
                  className="btn btn-primary"
                  onClick={() => submitEditStudio(editStudioModal)}
                  disabled={editStudioModal.isSubmitting}
                >
                  {editStudioModal.isSubmitting ? "Saving..." : "Save Changes"}
                </button>
              </>
            }
          >
            <div className="form-group">
              <label htmlFor="edit-studio-name">Studio Name *</label>
              <input
                id="edit-studio-name"
                type="text"
                className="form-input"
                placeholder="Enter studio name"
                {...vlens.attrsBindInput(
                  vlens.ref(editStudioModal, "studioName"),
                )}
                disabled={editStudioModal.isSubmitting}
              />
            </div>

            <div className="form-group">
              <label htmlFor="edit-studio-description">Description</label>
              <textarea
                id="edit-studio-description"
                className="form-input"
                placeholder="Enter studio description (optional)"
                rows={3}
                {...vlens.attrsBindInput(
                  vlens.ref(editStudioModal, "studioDescription"),
                )}
                disabled={editStudioModal.isSubmitting}
              />
            </div>

            <div className="form-group">
              <label htmlFor="edit-studio-max-rooms">Max Rooms *</label>
              <input
                id="edit-studio-max-rooms"
                type="number"
                min="1"
                className="form-input"
                {...vlens.attrsBindInput(
                  vlens.ref(editStudioModal, "maxRooms"),
                )}
                disabled={editStudioModal.isSubmitting}
              />
              <small className="form-help">
                Maximum number of streaming rooms allowed for this studio
              </small>
            </div>
          </Modal>

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
                {deleteStudioModal.isOwner && (
                  <button
                    className="btn btn-danger"
                    onClick={() => confirmDeleteStudio(deleteStudioModal)}
                    disabled={deleteStudioModal.isDeleting}
                  >
                    {deleteStudioModal.isDeleting
                      ? "Deleting..."
                      : "Delete Studio"}
                  </button>
                )}
              </>
            }
          >
            {!deleteStudioModal.isOwner ? (
              <div className="delete-warning">
                <p className="warning-text">
                  ⚠️ Only the studio owner can delete the studio.
                </p>
              </div>
            ) : (
              <div className="delete-confirmation">
                <p className="confirmation-text">
                  ⚠️ Are you sure you want to delete this studio? This will also
                  delete all rooms, streams, and memberships. This action cannot
                  be undone.
                </p>
                <div className="room-info">
                  <strong>Studio:</strong> {deleteStudioModal.studioName}
                </div>
              </div>
            )}
          </Modal>
        </div>
      </main>
      <Footer />
    </div>
  );
}
