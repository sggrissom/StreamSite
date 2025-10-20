import * as preact from "preact";
import * as vlens from "vlens";
import * as server from "../../../server";
import { Modal } from "../../../components/Modal";
import { Dropdown, DropdownItem } from "../../../components/Dropdown";
import { GenerateAccessCodeModal } from "./GenerateAccessCodeModal";

type Studio = {
  id: number;
  name: string;
  description: string;
  maxRooms: number;
};

type StudioHeaderProps = {
  studio: Studio;
  myRole: number;
  myRoleName: string;
  rooms: any[];
  members: any[];
  canManageRooms: boolean;
};

function getRoleBadgeClass(role: number): string {
  return `studio-role role-${role}`;
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

  closeDeleteStudioModal(modal);
  window.location.href = "/studios";
}

// ===== Generate Studio Access Code Modal State =====
type GenerateCodeModalState = {
  isOpen: boolean;
};

const useGenerateCodeModalState = vlens.declareHook(
  (): GenerateCodeModalState => ({
    isOpen: false,
  }),
);

function openGenerateCodeModal(modal: GenerateCodeModalState) {
  modal.isOpen = true;
  vlens.scheduleRedraw();
}

function closeGenerateCodeModal(modal: GenerateCodeModalState) {
  modal.isOpen = false;
  vlens.scheduleRedraw();
}

export function StudioHeader(props: StudioHeaderProps): preact.ComponentChild {
  const { studio, myRole, myRoleName, rooms, members, canManageRooms } = props;
  const editStudioModal = useEditStudioModal();
  const deleteStudioModal = useDeleteStudioModal();
  const generateCodeModal = useGenerateCodeModalState();

  const activeRooms = rooms.filter((r) => r.isActive).length;

  return (
    <>
      {/* Studio Header */}
      <div className="studio-header">
        <div className="studio-header-main">
          <h1 className="studio-title">{studio.name}</h1>
          <div className="studio-badges">
            <span className={getRoleBadgeClass(myRole)}>{myRoleName}</span>
            {activeRooms > 0 && (
              <span className="active-badge">🔴 {activeRooms} Live</span>
            )}
          </div>
        </div>
        {studio.description && (
          <p className="studio-description">{studio.description}</p>
        )}

        {/* Compact Stats */}
        <div className="studio-stats">
          <div className="studio-stats-text">
            <span className="stat-item">
              {rooms.length} / {studio.maxRooms} rooms
            </span>
            <span className="stat-separator">•</span>
            <span className="stat-item">{members.length} members</span>
          </div>
          {canManageRooms && (
            <Dropdown
              id={`studio-${studio.id}`}
              trigger={
                <button className="btn btn-secondary btn-sm">Actions ▼</button>
              }
              align="right"
            >
              <DropdownItem
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
              </DropdownItem>
              <DropdownItem
                onClick={() => openGenerateCodeModal(generateCodeModal)}
              >
                Generate Studio Access Code
              </DropdownItem>
              <DropdownItem
                onClick={() => {
                  document
                    .querySelector(".members-section")
                    ?.scrollIntoView({ behavior: "smooth" });
                }}
              >
                Manage Members
              </DropdownItem>
              {myRole === server.StudioRoleOwner && (
                <DropdownItem
                  onClick={() =>
                    openDeleteStudioModal(
                      deleteStudioModal,
                      studio.id,
                      studio.name,
                      true,
                    )
                  }
                  variant="danger"
                >
                  Delete Studio
                </DropdownItem>
              )}
            </Dropdown>
          )}
        </div>
      </div>

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
            {...vlens.attrsBindInput(vlens.ref(editStudioModal, "studioName"))}
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
            {...vlens.attrsBindInput(vlens.ref(editStudioModal, "maxRooms"))}
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
                {deleteStudioModal.isDeleting ? "Deleting..." : "Delete Studio"}
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
              delete all rooms, streams, and memberships. This action cannot be
              undone.
            </p>
            <div className="room-info">
              <strong>Studio:</strong> {deleteStudioModal.studioName}
            </div>
          </div>
        )}
      </Modal>

      {/* Generate Studio Access Code Modal */}
      <GenerateAccessCodeModal
        isOpen={generateCodeModal.isOpen}
        onClose={() => closeGenerateCodeModal(generateCodeModal)}
        codeType={1}
        targetId={studio.id}
        targetName={studio.name}
        targetLabel="Studio"
      />
    </>
  );
}
