import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
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

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  const roomModal = useCreateRoomModal();

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

          {/* Action Buttons Placeholder */}
          {canManageRooms && (
            <div className="studio-actions">
              <button className="btn btn-secondary" disabled>
                Edit Studio
              </button>
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
                          <button className="btn btn-secondary btn-sm" disabled>
                            View Stream Key
                          </button>
                          <button className="btn btn-secondary btn-sm" disabled>
                            Edit
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
          {roomModal.isOpen && (
            <div
              className="modal-overlay"
              onClick={() => closeRoomModal(roomModal)}
            >
              <div
                className="modal-content"
                onClick={(e) => e.stopPropagation()}
              >
                <div className="modal-header">
                  <h2 className="modal-title">Create New Room</h2>
                  <button
                    className="modal-close"
                    onClick={() => closeRoomModal(roomModal)}
                  >
                    ×
                  </button>
                </div>

                <div className="modal-body">
                  {roomModal.error && (
                    <div className="error-message">{roomModal.error}</div>
                  )}

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
                </div>

                <div className="modal-footer">
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
                </div>
              </div>
            </div>
          )}
        </div>
      </main>
      <Footer />
    </div>
  );
}
