import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../../server";
import { Modal } from "../../../components/Modal";
import { Dropdown, DropdownItem } from "../../../components/Dropdown";
import { GenerateAccessCodeModal } from "./GenerateAccessCodeModal";

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

// ===== Generate Access Code Modal State =====
type GenerateCodeModalState = {
  isOpen: boolean;
  roomId: number;
  roomName: string;
};

const useGenerateCodeModalState = vlens.declareHook(
  (): GenerateCodeModalState => ({
    isOpen: false,
    roomId: 0,
    roomName: "",
  }),
);

function openGenerateCodeModal(
  modal: GenerateCodeModalState,
  roomId: number,
  roomName: string,
) {
  modal.isOpen = true;
  modal.roomId = roomId;
  modal.roomName = roomName;
  vlens.scheduleRedraw();
}

function closeGenerateCodeModal(modal: GenerateCodeModalState) {
  modal.isOpen = false;
  vlens.scheduleRedraw();
}

// ===== View Analytics Modal =====
type ViewAnalyticsModal = {
  isOpen: boolean;
  isLoading: boolean;
  error: string;
  roomId: number;
  roomName: string;
  analytics: server.GetRoomAnalyticsResponse | null;
};

const useViewAnalyticsModal = vlens.declareHook(
  (): ViewAnalyticsModal => ({
    isOpen: false,
    isLoading: false,
    error: "",
    roomId: 0,
    roomName: "",
    analytics: null,
  }),
);

async function openAnalyticsModal(
  modal: ViewAnalyticsModal,
  roomId: number,
  roomName: string,
) {
  modal.isOpen = true;
  modal.isLoading = true;
  modal.error = "";
  modal.roomId = roomId;
  modal.roomName = roomName;
  modal.analytics = null;
  vlens.scheduleRedraw();

  const [resp, err] = await server.GetRoomAnalytics({ roomId });
  modal.isLoading = false;

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to load analytics";
    vlens.scheduleRedraw();
    return;
  }

  modal.analytics = resp;
  vlens.scheduleRedraw();
}

function closeAnalyticsModal(modal: ViewAnalyticsModal) {
  modal.isOpen = false;
  modal.error = "";
  modal.analytics = null;
  vlens.scheduleRedraw();
}

// ===== Component =====
export function RoomsSection(props: RoomsSectionProps): preact.ComponentChild {
  const { studio, rooms, canManageRooms } = props;
  const streamKeyModal = useStreamKeyModal();
  const editRoomModal = useEditRoomModal();
  const deleteRoomModal = useDeleteRoomModal();
  const generateCodeModal = useGenerateCodeModalState();
  const analyticsModal = useViewAnalyticsModal();

  return (
    <>
      <div className="rooms-section">
        <div className="rooms-header">
          <h2 className="section-title">Rooms</h2>
        </div>

        {rooms.length === 0 ? (
          <div className="rooms-empty">
            <div className="empty-icon">🎬</div>
            <h3>No Rooms Yet</h3>
            <p>
              {canManageRooms
                ? "Create your first room using the Actions dropdown above."
                : "This studio doesn't have any rooms yet."}
            </p>
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
                  <a
                    href={`/stream/${room.id}`}
                    className={`btn btn-sm ${room.isActive ? "btn-primary" : "btn-secondary"}`}
                  >
                    {room.isActive ? "Watch Stream" : "View Stream"}
                  </a>
                  {canManageRooms && (
                    <Dropdown
                      id={room.id}
                      trigger={
                        <button className="btn btn-secondary btn-sm">
                          Actions ▼
                        </button>
                      }
                      align="right"
                    >
                      <DropdownItem
                        onClick={() =>
                          openStreamKeyModal(streamKeyModal, room.id, room.name)
                        }
                      >
                        View Stream Key
                      </DropdownItem>
                      <DropdownItem
                        onClick={() =>
                          openAnalyticsModal(analyticsModal, room.id, room.name)
                        }
                      >
                        View Analytics
                      </DropdownItem>
                      <DropdownItem
                        onClick={() =>
                          openGenerateCodeModal(
                            generateCodeModal,
                            room.id,
                            room.name,
                          )
                        }
                      >
                        Generate Access Code
                      </DropdownItem>
                      <DropdownItem
                        onClick={() =>
                          openEditRoomModal(editRoomModal, room.id, room.name)
                        }
                      >
                        Edit Room
                      </DropdownItem>
                      <DropdownItem
                        onClick={() =>
                          openDeleteRoomModal(
                            deleteRoomModal,
                            room.id,
                            room.name,
                            room.isActive,
                          )
                        }
                        variant="danger"
                      >
                        Delete Room
                      </DropdownItem>
                    </Dropdown>
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

      <GenerateAccessCodeModal
        isOpen={generateCodeModal.isOpen}
        onClose={() => closeGenerateCodeModal(generateCodeModal)}
        codeType={0}
        targetId={generateCodeModal.roomId}
        targetName={generateCodeModal.roomName}
        targetLabel="Room"
      />

      <Modal
        isOpen={analyticsModal.isOpen}
        title="Room Analytics"
        onClose={() => closeAnalyticsModal(analyticsModal)}
        error={analyticsModal.error}
        footer={
          <button
            className="btn btn-secondary"
            onClick={() => closeAnalyticsModal(analyticsModal)}
          >
            Close
          </button>
        }
      >
        <div className="form-group">
          <label>Room</label>
          <div style="font-weight: 500;">{analyticsModal.roomName}</div>
        </div>

        {analyticsModal.isLoading ? (
          <div style="padding: 1rem; text-align: center;">Loading...</div>
        ) : (
          analyticsModal.analytics &&
          analyticsModal.analytics.analytics && (
            <>
              <div className="form-group">
                <label>Current Viewers</label>
                <div style="font-size: 1.5rem; font-weight: bold; color: var(--primary);">
                  {analyticsModal.analytics.analytics.currentViewers}
                </div>
              </div>

              <div className="form-group">
                <label>Peak Viewers</label>
                <div style="font-size: 1.25rem; font-weight: 600;">
                  {analyticsModal.analytics.analytics.peakViewers}
                  {analyticsModal.analytics.analytics.peakViewersAt && (
                    <small style="display: block; font-size: 0.875rem; font-weight: normal; color: var(--text-secondary); margin-top: 0.25rem;">
                      {new Date(
                        analyticsModal.analytics.analytics.peakViewersAt,
                      ).toLocaleString()}
                    </small>
                  )}
                </div>
              </div>

              <div className="form-group">
                <label>Total Views</label>
                <div style="font-size: 1.125rem; font-weight: 600;">
                  All-time:{" "}
                  {analyticsModal.analytics.analytics.totalViewsAllTime}
                  <br />
                  This month:{" "}
                  {analyticsModal.analytics.analytics.totalViewsThisMonth}
                </div>
              </div>

              <div className="form-group">
                <label>Streaming Status</label>
                <div>
                  {analyticsModal.analytics.isStreaming ? (
                    <span style="color: var(--success); font-weight: 600;">
                      🔴 Live Now
                    </span>
                  ) : (
                    <span style="color: var(--text-secondary);">Offline</span>
                  )}
                </div>
              </div>

              <div className="form-group">
                <label>Total Stream Time</label>
                <div style="font-weight: 500;">
                  {Math.floor(
                    analyticsModal.analytics.analytics.totalStreamMinutes / 60,
                  )}{" "}
                  hours{" "}
                  {analyticsModal.analytics.analytics.totalStreamMinutes % 60}{" "}
                  minutes
                </div>
              </div>

              {analyticsModal.analytics.analytics.lastStreamAt && (
                <div className="form-group">
                  <label>Last Streamed</label>
                  <div style="font-weight: 500;">
                    {new Date(
                      analyticsModal.analytics.analytics.lastStreamAt,
                    ).toLocaleString()}
                  </div>
                </div>
              )}
            </>
          )
        )}
      </Modal>
    </>
  );
}
