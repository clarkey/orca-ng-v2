import * as React from "react";
import { UseFormReturn } from "react-hook-form";
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "./form";
import { Input } from "./input";
import { Checkbox } from "./checkbox";

// ============================================================================
// Standard Form Checkbox - Reusable pattern we've established
// ============================================================================

interface FormCheckboxProps {
  form: UseFormReturn<any>;
  name: string;
  label: string;
  description?: string;
}

export function FormCheckbox({ form, name, label, description }: FormCheckboxProps) {
  const fieldId = React.useId();
  
  return (
    <FormField
      control={form.control}
      name={name}
      render={({ field }) => (
        <FormItem className="flex flex-row items-start space-x-3 space-y-0">
          <FormControl>
            <Checkbox
              id={fieldId}
              checked={field.value}
              onCheckedChange={field.onChange}
            />
          </FormControl>
          <label 
            htmlFor={fieldId}
            className="space-y-1 cursor-pointer block -mt-1"
          >
            <div className="text-sm font-medium text-gray-700">
              {label}
            </div>
            {description && (
              <FormDescription>
                {description}
              </FormDescription>
            )}
          </label>
        </FormItem>
      )}
    />
  );
}

// ============================================================================
// Standard Form Input - Text, Email, Password, URL, etc.
// ============================================================================

interface FormInputProps {
  form: UseFormReturn<any>;
  name: string;
  label?: string;
  placeholder?: string;
  description?: string;
  type?: "text" | "email" | "password" | "url" | "number";
  autoComplete?: string;
  disabled?: boolean;
}

export function FormInput({ 
  form, 
  name, 
  label, 
  placeholder,
  description,
  type = "text",
  autoComplete,
  disabled
}: FormInputProps) {
  return (
    <FormField
      control={form.control}
      name={name}
      render={({ field }) => (
        <FormItem>
          {label && <FormLabel>{label}</FormLabel>}
          <FormControl>
            <Input
              type={type}
              placeholder={placeholder}
              autoComplete={autoComplete}
              disabled={disabled}
              {...field}
            />
          </FormControl>
          {description && <FormDescription>{description}</FormDescription>}
          <FormMessage />
        </FormItem>
      )}
    />
  );
}

