import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../../server";
import { Modal } from "../../../components/Modal";
import { Dropdown, DropdownItem } from "../../../components/Dropdown";
import qrcode from "qrcode-generator";

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

// ===== Generate Access Code Modal =====
type GenerateCodeModal = {
  isOpen: boolean;
  isSubmitting: boolean;
  error: string;
  roomId: number;
  roomName: string;
  duration: string;
  maxViewers: number;
  label: string;
  generatedCode: string;
  shareUrl: string;
  qrDataUrl: string;
  showSuccess: boolean;
  copyCodeSuccess: boolean;
  copyUrlSuccess: boolean;
};

const useGenerateCodeModal = vlens.declareHook(
  (): GenerateCodeModal => ({
    isOpen: false,
    isSubmitting: false,
    error: "",
    roomId: 0,
    roomName: "",
    duration: "24h",
    maxViewers: 30,
    label: "",
    generatedCode: "",
    shareUrl: "",
    qrDataUrl: "",
    showSuccess: false,
    copyCodeSuccess: false,
    copyUrlSuccess: false,
  }),
);

function openGenerateCodeModal(
  modal: GenerateCodeModal,
  roomId: number,
  roomName: string,
) {
  modal.isOpen = true;
  modal.error = "";
  modal.roomId = roomId;
  modal.roomName = roomName;
  modal.duration = "24h";
  modal.maxViewers = 30;
  modal.label = "";
  modal.generatedCode = "";
  modal.shareUrl = "";
  modal.qrDataUrl = "";
  modal.showSuccess = false;
  modal.copyCodeSuccess = false;
  modal.copyUrlSuccess = false;
  vlens.scheduleRedraw();
}

function closeGenerateCodeModal(modal: GenerateCodeModal) {
  modal.isOpen = false;
  modal.error = "";
  modal.showSuccess = false;
  vlens.scheduleRedraw();
}

function durationToMinutes(duration: string): number {
  switch (duration) {
    case "1h":
      return 60;
    case "24h":
      return 1440;
    case "7d":
      return 10080;
    case "30d":
      return 43200;
    case "never":
      return -1;
    default:
      return 1440; // default to 24h
  }
}

async function submitGenerateCode(modal: GenerateCodeModal) {
  if (modal.maxViewers < 1) {
    modal.error = "Max viewers must be at least 1";
    vlens.scheduleRedraw();
    return;
  }

  modal.isSubmitting = true;
  modal.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.GenerateAccessCode({
    type: 0, // CodeTypeRoom
    targetId: modal.roomId,
    durationMinutes: durationToMinutes(modal.duration),
    maxViewers: modal.maxViewers,
    label: modal.label.trim() || "",
  });

  modal.isSubmitting = false;

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to generate access code";
    vlens.scheduleRedraw();
    return;
  }

  // Show success view with generated code
  modal.generatedCode = resp.code || "";
  const hostname =
    window.location.hostname === "localhost"
      ? "localhost:3000"
      : "stream.grissom.zone";
  modal.shareUrl = `http://${hostname}/watch/${resp.code}`;

  // Generate QR code for the share URL
  try {
    const qr = qrcode(0, "M"); // Type number (0 = auto), error correction level 'M'
    qr.addData(modal.shareUrl);
    qr.make();
    modal.qrDataUrl = qr.createDataURL(4); // Cell size in pixels
  } catch (qrError) {
    console.error("Failed to generate QR code:", qrError);
    modal.qrDataUrl = ""; // Continue without QR code if it fails
  }

  modal.showSuccess = true;
  vlens.scheduleRedraw();
}

async function copyCode(modal: GenerateCodeModal) {
  try {
    await navigator.clipboard.writeText(modal.generatedCode);
    modal.copyCodeSuccess = true;
    vlens.scheduleRedraw();
    setTimeout(() => {
      modal.copyCodeSuccess = false;
      vlens.scheduleRedraw();
    }, 2000);
  } catch (err) {
    modal.error = "Failed to copy to clipboard";
    vlens.scheduleRedraw();
  }
}

async function copyShareUrl(modal: GenerateCodeModal) {
  try {
    await navigator.clipboard.writeText(modal.shareUrl);
    modal.copyUrlSuccess = true;
    vlens.scheduleRedraw();
    setTimeout(() => {
      modal.copyUrlSuccess = false;
      vlens.scheduleRedraw();
    }, 2000);
  } catch (err) {
    modal.error = "Failed to copy to clipboard";
    vlens.scheduleRedraw();
  }
}

function downloadQRCode(modal: GenerateCodeModal) {
  if (!modal.qrDataUrl) return;

  // Create a temporary link element to trigger download
  const link = document.createElement("a");
  link.href = modal.qrDataUrl;
  link.download = `access-code-${modal.generatedCode}.png`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

// ===== Component =====
export function RoomsSection(props: RoomsSectionProps): preact.ComponentChild {
  const { studio, rooms, canManageRooms } = props;
  const roomModal = useCreateRoomModal();
  const streamKeyModal = useStreamKeyModal();
  const editRoomModal = useEditRoomModal();
  const deleteRoomModal = useDeleteRoomModal();
  const generateCodeModal = useGenerateCodeModal();

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

      <Modal
        isOpen={generateCodeModal.isOpen}
        title={
          generateCodeModal.showSuccess
            ? "Access Code Generated"
            : "Generate Access Code"
        }
        onClose={() => closeGenerateCodeModal(generateCodeModal)}
        error={generateCodeModal.error}
        footer={
          generateCodeModal.showSuccess ? (
            <button
              className="btn btn-primary"
              onClick={() => closeGenerateCodeModal(generateCodeModal)}
            >
              Done
            </button>
          ) : (
            <>
              <button
                className="btn btn-secondary"
                onClick={() => closeGenerateCodeModal(generateCodeModal)}
                disabled={generateCodeModal.isSubmitting}
              >
                Cancel
              </button>
              <button
                className="btn btn-primary"
                onClick={() => submitGenerateCode(generateCodeModal)}
                disabled={generateCodeModal.isSubmitting}
              >
                {generateCodeModal.isSubmitting
                  ? "Generating..."
                  : "Generate Code"}
              </button>
            </>
          )
        }
      >
        {generateCodeModal.showSuccess ? (
          // Success view - show generated code and share URL
          <div>
            <div className="form-group">
              <label>Room</label>
              <div className="stream-key-room-name">
                {generateCodeModal.roomName}
              </div>
            </div>

            <div className="form-group">
              <label>Access Code</label>
              <div
                className="stream-key-display"
                style="font-size: 2rem; letter-spacing: 0.5rem; text-align: center; padding: 1.5rem;"
              >
                {generateCodeModal.generatedCode}
              </div>
              <small className="form-help">
                Share this 5-digit code with viewers to grant them access
              </small>
            </div>

            <div className="form-group">
              <label>Share URL</label>
              <div className="stream-key-display">
                {generateCodeModal.shareUrl}
              </div>
              <small className="form-help">
                Direct link that viewers can use to access the stream
              </small>
            </div>

            {generateCodeModal.qrDataUrl && (
              <div className="form-group">
                <label>QR Code</label>
                <div style="text-align: center; padding: 1rem; background: white; border: 1px solid var(--border); border-radius: 8px;">
                  <img
                    src={generateCodeModal.qrDataUrl}
                    alt="QR Code for access"
                    style="max-width: 256px; width: 100%; height: auto;"
                  />
                </div>
                <small className="form-help">
                  Scan this QR code to quickly access the stream
                </small>
              </div>
            )}

            <div className="stream-key-actions">
              <button
                className="btn btn-primary"
                onClick={() => copyCode(generateCodeModal)}
                disabled={generateCodeModal.copyCodeSuccess}
              >
                {generateCodeModal.copyCodeSuccess
                  ? "✓ Copied Code!"
                  : "Copy Code"}
              </button>
              <button
                className="btn btn-secondary"
                onClick={() => copyShareUrl(generateCodeModal)}
                disabled={generateCodeModal.copyUrlSuccess}
              >
                {generateCodeModal.copyUrlSuccess
                  ? "✓ Copied URL!"
                  : "Copy URL"}
              </button>
              {generateCodeModal.qrDataUrl && (
                <button
                  className="btn btn-secondary"
                  onClick={() => downloadQRCode(generateCodeModal)}
                >
                  Download QR Code
                </button>
              )}
            </div>
          </div>
        ) : (
          // Form view - collect code parameters
          <div>
            <div className="form-group">
              <label>Room</label>
              <div className="stream-key-room-name">
                {generateCodeModal.roomName}
              </div>
            </div>

            <div className="form-group">
              <label htmlFor="code-duration">Code Duration *</label>
              <select
                id="code-duration"
                className="form-input"
                {...vlens.attrsBindInput(
                  vlens.ref(generateCodeModal, "duration"),
                )}
                disabled={generateCodeModal.isSubmitting}
              >
                <option value="1h">1 hour</option>
                <option value="24h">24 hours</option>
                <option value="7d">7 days</option>
                <option value="30d">30 days</option>
                <option value="never">Never expires</option>
              </select>
              <small className="form-help">
                How long the access code will remain valid
              </small>
            </div>

            <div className="form-group">
              <label htmlFor="code-max-viewers">Max Viewers *</label>
              <input
                id="code-max-viewers"
                type="number"
                className="form-input"
                min="1"
                placeholder="e.g., 30"
                {...vlens.attrsBindInput(
                  vlens.ref(generateCodeModal, "maxViewers"),
                )}
                disabled={generateCodeModal.isSubmitting}
              />
              <small className="form-help">
                Maximum number of viewers who can use this code
              </small>
            </div>

            <div className="form-group">
              <label htmlFor="code-label">Label (Optional)</label>
              <input
                id="code-label"
                type="text"
                className="form-input"
                placeholder="e.g., Class Section A, Guest Viewers"
                {...vlens.attrsBindInput(vlens.ref(generateCodeModal, "label"))}
                disabled={generateCodeModal.isSubmitting}
              />
              <small className="form-help">
                Optional label to help you identify this code later
              </small>
            </div>
          </div>
        )}
      </Modal>
    </>
  );
}
