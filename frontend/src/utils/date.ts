/**
 * Formats a UTC date string to the user's local timezone
 * @param dateString - ISO 8601 date string from the API (in UTC)
 * @param options - Intl.DateTimeFormatOptions for formatting
 * @returns Formatted date string in user's local timezone
 */
export function formatDate(
  dateString: string | undefined | null,
  options: Intl.DateTimeFormatOptions = {
    dateStyle: 'medium',
    timeStyle: 'short',
  }
): string {
  if (!dateString) return 'Never';
  
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return 'Invalid date';
    
    return new Intl.DateTimeFormat(undefined, options).format(date);
  } catch (error) {
    console.error('Error formatting date:', error);
    return 'Invalid date';
  }
}

/**
 * Formats a date as relative time (e.g., "2 hours ago", "in 5 minutes")
 * @param dateString - ISO 8601 date string from the API (in UTC)
 * @returns Relative time string
 */
export function formatRelativeTime(dateString: string | undefined | null): string {
  if (!dateString) return 'Never';
  
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return 'Invalid date';
    
    const now = new Date();
    const diffInMs = now.getTime() - date.getTime();
    const diffInMinutes = Math.floor(diffInMs / 60000);
    const diffInHours = Math.floor(diffInMinutes / 60);
    const diffInDays = Math.floor(diffInHours / 24);
    
    if (diffInMinutes < 1) return 'Just now';
    if (diffInMinutes < 60) return `${diffInMinutes} minute${diffInMinutes !== 1 ? 's' : ''} ago`;
    if (diffInHours < 24) return `${diffInHours} hour${diffInHours !== 1 ? 's' : ''} ago`;
    if (diffInDays < 30) return `${diffInDays} day${diffInDays !== 1 ? 's' : ''} ago`;
    
    // Fall back to absolute date for older dates
    return formatDate(dateString, { dateStyle: 'medium' });
  } catch (error) {
    console.error('Error formatting relative time:', error);
    return 'Invalid date';
  }
}

/**
 * Gets the user's timezone
 * @returns The user's timezone string (e.g., "America/New_York")
 */
export function getUserTimezone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}

/**
 * Checks if a session is expired
 * @param expiresAt - ISO 8601 expiry date string from the API (in UTC)
 * @returns true if the session is expired
 */
export function isSessionExpired(expiresAt: string | undefined | null): boolean {
  if (!expiresAt) return true;
  
  try {
    const expiryDate = new Date(expiresAt);
    return new Date() > expiryDate;
  } catch (error) {
    console.error('Error checking session expiry:', error);
    return true;
  }
}