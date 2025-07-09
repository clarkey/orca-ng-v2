/**
 * Shared UI constants for consistent styling across components
 */

// Overlay styles used for both Dialog and AlertDialog
export const overlayStyles = {
  base: "fixed inset-0 z-50 bg-black/80",
  animation: "data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
} as const;

// Modal content positioning and animation
export const modalContentStyles = {
  position: "fixed left-[50%] top-[50%] z-50 translate-x-[-50%] translate-y-[-50%]",
  appearance: "bg-white shadow-lg",
  animation: "duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%] data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%]",
  // Dialog specific - Updated to handle overflow properly
  dialog: "grid w-full max-w-lg gap-4 border p-6 sm:rounded max-h-[90vh] overflow-hidden",
  // AlertDialog specific
  alertDialog: "grid w-full max-w-lg gap-4 border p-6 sm:rounded-lg max-h-[90vh] overflow-hidden",
} as const;

// Combine overlay styles for easy reuse
export const getOverlayClassName = () => 
  `${overlayStyles.base} ${overlayStyles.animation}`;

// Get dialog content styles
export const getDialogContentClassName = () => 
  `${modalContentStyles.position} ${modalContentStyles.appearance} ${modalContentStyles.animation} ${modalContentStyles.dialog}`;

// Get alert dialog content styles
export const getAlertDialogContentClassName = () => 
  `${modalContentStyles.position} ${modalContentStyles.appearance} ${modalContentStyles.animation} ${modalContentStyles.alertDialog}`;