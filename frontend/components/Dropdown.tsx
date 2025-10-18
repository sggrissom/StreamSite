import * as preact from "preact";
import * as vlens from "vlens";
import { clickOutside } from "vlens/core";

type DropdownProps = {
  id: string | number;
  trigger: preact.ComponentChild;
  children: preact.ComponentChild;
  align?: "left" | "right";
};

type DropdownState = {
  isOpen: boolean;
  dropdownRef: HTMLDivElement | null;
  onRefSet: (el: HTMLDivElement | null) => void;
};

const useDropdown = vlens.declareHook((id: string | number): DropdownState => {
  const state: DropdownState = {
    isOpen: false,
    dropdownRef: null,
    onRefSet: (el: HTMLDivElement | null) => {
      state.dropdownRef = el;
    },
  };
  return state;
});

function toggleDropdown(state: DropdownState, event: Event) {
  event.stopPropagation();
  state.isOpen = !state.isOpen;
  vlens.scheduleRedraw();
}

function closeDropdown(state: DropdownState) {
  if (state.isOpen) {
    state.isOpen = false;
    vlens.scheduleRedraw();
  }
}

export function Dropdown(props: DropdownProps): preact.ComponentChild {
  const { id, trigger, children, align = "left" } = props;
  const state = useDropdown(id);

  // Close dropdown when clicking outside
  if (clickOutside(state.dropdownRef)) {
    closeDropdown(state);
  }

  return (
    <div className="dropdown" ref={state.onRefSet}>
      <div
        className="dropdown-trigger"
        onClick={(e) => toggleDropdown(state, e)}
      >
        {trigger}
      </div>
      {state.isOpen && (
        <div className={`dropdown-menu dropdown-menu-${align}`}>
          <div onClick={() => closeDropdown(state)}>{children}</div>
        </div>
      )}
    </div>
  );
}

type DropdownItemProps = {
  onClick: () => void;
  children: preact.ComponentChild;
  variant?: "default" | "danger";
};

export function DropdownItem(props: DropdownItemProps): preact.ComponentChild {
  const { onClick, children, variant = "default" } = props;

  return (
    <button
      className={`dropdown-item dropdown-item-${variant}`}
      onClick={(e) => {
        e.stopPropagation();
        onClick();
      }}
    >
      {children}
    </button>
  );
}
