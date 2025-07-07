import * as React from "react";
import { DialogTitle, DialogDescription } from "./dialog";
import { cn } from "@/lib/utils";

interface DialogHeaderStyledProps {
  title: string;
  description?: string;
  className?: string;
  children?: React.ReactNode;
}

export function DialogHeaderStyled({ 
  title, 
  description, 
  className,
  children 
}: DialogHeaderStyledProps) {
  return (
    <div className={cn("bg-gray-50 border-b border-gray-200 px-6 py-5", className)}>
      <div className="flex flex-col space-y-1 text-center sm:text-left">
        <DialogTitle className="text-lg font-semibold text-gray-900">
          {title}
        </DialogTitle>
        {description && (
          <DialogDescription className="text-sm text-gray-600">
            {description}
          </DialogDescription>
        )}
      </div>
      {children}
    </div>
  );
}