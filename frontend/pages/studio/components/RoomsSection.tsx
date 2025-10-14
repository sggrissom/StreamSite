import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../../server";
import { Modal } from "../../../components/Modal";

type Studio = {
  id: number;
  name: string;
  maxRooms: number;
};

type Room = {
  id: number;
  name: string;
  roomNumber: number;
  isActive: boolean;
  creation: string;
};

type RoomsSectionProps = {
  studio: Studio;
  rooms: Room[];
  canManageRooms: boolean;
};

// ===== Create Room Modal =====
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

  closeRoomModal(modal);
  window.location.reload();
}

// ===== Stream Key Modal =====
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
    // Build complete RTMP URL with stream key
    const hostname =
      window.location.hostname === "localhost"
        ? "localhost"
        : "stream.grissom.zone";
    const completeUrl = `rtmp://${hostname}:1935/live/${modal.streamKey}`;

    await navigator.clipboard.writeText(completeUrl);
    modal.copySuccess = true;
    vlens.scheduleRedraw();
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

// ===== Edit Room Modal =====
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

  closeEditRoomModal(modal);
  window.location.reload();
}

// ===== Delete Room Modal =====
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

  const [resp, err] = await server.DeleteRoom({ roomId: modal.roomId });
  modal.isDeleting = false;

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to delete room";
    vlens.scheduleRedraw();
    return;
  }

  closeDeleteRoomModal(modal);
  window.location.reload();
}

// ===== Component =====
export function RoomsSection(props: RoomsSectionProps): preact.ComponentChild {
  const { studio, rooms, canManageRooms } = props;
  const roomModal = useCreateRoomModal();
  const streamKeyModal = useStreamKeyModal();
  const editRoomModal = useEditRoomModal();
  const deleteRoomModal = useDeleteRoomModal();

  return (
    <>
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
                  {room.isActive && (
                    <a
                      href={`/stream/${room.id}`}
                      className="btn btn-primary btn-sm"
                    >
                      Watch Stream
                    </a>
                  )}
                  {canManageRooms && (
                    <>
                      <button
                        className="btn btn-secondary btn-sm"
                        onClick={() =>
                          openStreamKeyModal(streamKeyModal, room.id, room.name)
                        }
                      >
                        View Stream Key
                      </button>
                      <button
                        className="btn btn-secondary btn-sm"
                        onClick={() =>
                          openEditRoomModal(editRoomModal, room.id, room.name)
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
                </div>
              </div>
            ))}
          </div>
        )}

        {canManageRooms && rooms.length >= studio.maxRooms && (
          <div className="rooms-limit-notice">
            <p>
              ⚠️ You've reached the maximum number of rooms ({studio.maxRooms}).
              To create more rooms, increase the limit in studio settings.
            </p>
          </div>
        )}
      </div>

      {/* Modals */}
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
          <div className="stream-key-room-name">{streamKeyModal.roomName}</div>
        </div>

        {streamKeyModal.isLoading ? (
          <div className="stream-key-loading">Loading...</div>
        ) : (
          streamKeyModal.streamKey && (
            <>
              <div className="form-group">
                <label>RTMP Server URL</label>
                <div className="stream-key-display">
                  rtmp://
                  {window.location.hostname === "localhost"
                    ? "localhost"
                    : "stream.grissom.zone"}
                  :1935/live
                </div>
                <small className="form-help">
                  Enter this URL in your streaming software's server field
                </small>
              </div>

              <div className="form-group">
                <label>Stream Key</label>
                <div className="stream-key-display">
                  {streamKeyModal.streamKey}
                </div>
                <small className="form-help">
                  Enter this key in your streaming software's stream key field
                </small>
              </div>

              <div className="stream-key-actions">
                <button
                  className="btn btn-primary"
                  onClick={() => copyStreamKey(streamKeyModal)}
                  disabled={streamKeyModal.copySuccess}
                >
                  {streamKeyModal.copySuccess
                    ? "✓ Copied URL!"
                    : "Copy Complete URL"}
                </button>
                <button
                  className="btn btn-secondary"
                  onClick={() => showRegenerateConfirmation(streamKeyModal)}
                  disabled={streamKeyModal.isRegenerating}
                >
                  Regenerate Key
                </button>
              </div>
              <small
                className="form-help"
                style="margin-top: 0.5rem; display: block;"
              >
                The "Copy Complete URL" button copies the full RTMP URL with
                your stream key included, ready to paste into streaming software
                like Larix.
              </small>

              {streamKeyModal.showConfirmRegenerate && (
                <div className="confirmation-dialog">
                  <p className="confirmation-text">
                    ⚠️ Are you sure you want to regenerate the stream key? The
                    old key will stop working immediately.
                  </p>
                  <div className="confirmation-actions">
                    <button
                      className="btn btn-secondary btn-sm"
                      onClick={() => hideRegenerateConfirmation(streamKeyModal)}
                      disabled={streamKeyModal.isRegenerating}
                    >
                      Cancel
                    </button>
                    <button
                      className="btn btn-danger btn-sm"
                      onClick={() => confirmRegenerateStreamKey(streamKeyModal)}
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
              ⚠️ Cannot delete this room while it is actively streaming. Please
              stop the stream first.
            </p>
            <div className="room-info">
              <strong>Room:</strong> {deleteRoomModal.roomName}
            </div>
          </div>
        ) : (
          <div className="delete-confirmation">
            <p className="confirmation-text">
              ⚠️ Are you sure you want to delete this room? This action cannot
              be undone.
            </p>
            <div className="room-info">
              <strong>Room:</strong> {deleteRoomModal.roomName}
            </div>
          </div>
        )}
      </Modal>
    </>
  );
}
