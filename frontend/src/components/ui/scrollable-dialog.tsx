import * as React from "react"
import * as DialogPrimitive from "@radix-ui/react-dialog"
import { X } from "lucide-react"

import { cn } from "@/lib/utils"
import { Dialog, DialogPortal, DialogOverlay } from "./dialog"

const ScrollableDialogContent = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content> & {
    hideCloseButton?: boolean;
  }
>(({ className, children, hideCloseButton = false, ...props }, ref) => (
  <DialogPortal>
    <DialogOverlay />
    <DialogPrimitive.Content
      ref={ref}
      className={cn(
        // Base positioning and appearance
        "fixed left-[50%] top-[50%] z-50 w-full max-w-lg translate-x-[-50%] translate-y-[-50%]",
        "bg-white shadow-lg sm:rounded-lg",
        // Set max height and enable flex layout
        "max-h-[90vh] flex flex-col",
        // Animation
        "duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out",
        "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
        "data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95",
        "data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%]",
        "data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%]",
        className
      )}
      {...props}
    >
      {/* Content wrapper that enables scrolling */}
      <div className="flex flex-col max-h-[90vh] overflow-hidden">
        {children}
      </div>
      {!hideCloseButton && (
        <DialogPrimitive.Close className="absolute right-4 top-4 rounded opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground">
          <X className="h-4 w-4" />
          <span className="sr-only">Close</span>
        </DialogPrimitive.Close>
      )}
    </DialogPrimitive.Content>
  </DialogPortal>
))
ScrollableDialogContent.displayName = "ScrollableDialogContent"

// Header component that stays fixed at the top
const ScrollableDialogHeader = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    className={cn(
      "flex-shrink-0 px-6 pt-6 pb-4",
      className
    )}
    {...props}
  />
)
ScrollableDialogHeader.displayName = "ScrollableDialogHeader"

// Body component that scrolls
const ScrollableDialogBody = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    className={cn(
      "flex-1 overflow-y-auto px-6 pb-4",
      // Add some padding at the bottom for better scroll experience
      "scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100",
      className
    )}
    {...props}
  />
)
ScrollableDialogBody.displayName = "ScrollableDialogBody"

// Footer component that stays fixed at the bottom
const ScrollableDialogFooter = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    className={cn(
      "flex-shrink-0 px-6 py-4 border-t",
      "flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2",
      className
    )}
    {...props}
  />
)
ScrollableDialogFooter.displayName = "ScrollableDialogFooter"

export {
  ScrollableDialogContent,
  ScrollableDialogHeader,
  ScrollableDialogBody,
  ScrollableDialogFooter,
}