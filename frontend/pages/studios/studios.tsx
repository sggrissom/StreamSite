import * as preact from "preact";
import * as vlens from "vlens";
import * as core from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./studios-styles";

type Data = server.ListMyStudiosResponse;

export async function fetch(route: string, prefix: string) {
  return server.ListMyStudios({});
}

type CreateStudioModal = {
  isOpen: boolean;
  isSubmitting: boolean;
  error: string;
  name: string;
  description: string;
  maxRooms: string;
};

const useCreateStudioModal = vlens.declareHook(
  (): CreateStudioModal => ({
    isOpen: false,
    isSubmitting: false,
    error: "",
    name: "",
    description: "",
    maxRooms: "5",
  }),
);

function openModal(modal: CreateStudioModal) {
  modal.isOpen = true;
  modal.error = "";
  modal.name = "";
  modal.description = "";
  modal.maxRooms = "5";
  vlens.scheduleRedraw();
}

function closeModal(modal: CreateStudioModal) {
  modal.isOpen = false;
  modal.error = "";
  vlens.scheduleRedraw();
}

async function submitCreateStudio(modal: CreateStudioModal) {
  if (!modal.name.trim()) {
    modal.error = "Studio name is required";
    vlens.scheduleRedraw();
    return;
  }

  const maxRooms = parseInt(modal.maxRooms);
  if (isNaN(maxRooms) || maxRooms < 1 || maxRooms > 100) {
    modal.error = "Max rooms must be between 1 and 100";
    vlens.scheduleRedraw();
    return;
  }

  modal.isSubmitting = true;
  modal.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.CreateStudio({
    name: modal.name.trim(),
    description: modal.description.trim(),
    maxRooms: maxRooms,
  });

  modal.isSubmitting = false;

  if (err || !resp || !resp.success) {
    modal.error = resp?.error || err || "Failed to create studio";
    vlens.scheduleRedraw();
    return;
  }

  // Success - close modal and reload page
  closeModal(modal);
  window.location.reload();
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  const studios = data?.studios || [];
  const modal = useCreateStudioModal();

  return (
    <div>
      <Header />
      <main className="studios-container">
        <div className="studios-content">
          <div className="studios-header">
            <h1 className="studios-title">My Studios</h1>
            <p className="studios-description">
              Studios you are a member of. Create or join studios to start
              streaming.
            </p>
          </div>

          {studios.length === 0 ? (
            <div className="studios-empty">
              <div className="empty-icon">📺</div>
              <h2>No Studios Yet</h2>
              <p>
                You haven't joined any studios yet. Create your first studio to
                get started!
              </p>
              <button
                className="btn btn-primary"
                onClick={() => openModal(modal)}
              >
                Create Your First Studio
              </button>
            </div>
          ) : (
            <div className="studios-grid">
              {studios.map((studio) => (
                <div key={studio.id} className="studio-card">
                  <div className="studio-card-header">
                    <h3 className="studio-name">{studio.name}</h3>
                    <span className={`studio-role role-${studio.myRole}`}>
                      {studio.myRoleName}
                    </span>
                  </div>

                  {studio.description && (
                    <p className="studio-description">{studio.description}</p>
                  )}

                  <div className="studio-meta">
                    <span className="meta-item">
                      <span className="meta-label">Max Rooms:</span>
                      <span className="meta-value">{studio.maxRooms}</span>
                    </span>
                  </div>

                  <div className="studio-actions">
                    <a
                      href={`/studio/${studio.id}`}
                      className="btn btn-primary btn-sm"
                    >
                      Open Studio
                    </a>
                  </div>
                </div>
              ))}
            </div>
          )}

          {studios.length > 0 && (
            <div className="studios-footer">
              <button
                className="btn btn-primary"
                onClick={() => openModal(modal)}
              >
                Create New Studio
              </button>
            </div>
          )}

          {modal.isOpen && (
            <div className="modal-overlay" onClick={() => closeModal(modal)}>
              <div
                className="modal-content"
                onClick={(e) => e.stopPropagation()}
              >
                <div className="modal-header">
                  <h2 className="modal-title">Create New Studio</h2>
                  <button
                    className="modal-close"
                    onClick={() => closeModal(modal)}
                  >
                    ×
                  </button>
                </div>

                <div className="modal-body">
                  {modal.error && (
                    <div className="error-message">{modal.error}</div>
                  )}

                  <div className="form-group">
                    <label htmlFor="studio-name">Studio Name *</label>
                    <input
                      id="studio-name"
                      type="text"
                      className="form-input"
                      placeholder="Enter studio name"
                      {...vlens.attrsBindInput(vlens.ref(modal, "name"))}
                      disabled={modal.isSubmitting}
                    />
                  </div>

                  <div className="form-group">
                    <label htmlFor="studio-description">Description</label>
                    <textarea
                      id="studio-description"
                      className="form-input"
                      placeholder="Enter studio description (optional)"
                      rows={3}
                      {...vlens.attrsBindInput(vlens.ref(modal, "description"))}
                      disabled={modal.isSubmitting}
                    />
                  </div>

                  <div className="form-group">
                    <label htmlFor="studio-max-rooms">Max Rooms *</label>
                    <input
                      id="studio-max-rooms"
                      type="number"
                      className="form-input"
                      min="1"
                      max="100"
                      {...vlens.attrsBindInput(vlens.ref(modal, "maxRooms"))}
                      disabled={modal.isSubmitting}
                    />
                    <small className="form-help">
                      Maximum number of streaming rooms (1-100)
                    </small>
                  </div>
                </div>

                <div className="modal-footer">
                  <button
                    className="btn btn-secondary"
                    onClick={() => closeModal(modal)}
                    disabled={modal.isSubmitting}
                  >
                    Cancel
                  </button>
                  <button
                    className="btn btn-primary"
                    onClick={() => submitCreateStudio(modal)}
                    disabled={modal.isSubmitting}
                  >
                    {modal.isSubmitting ? "Creating..." : "Create Studio"}
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
