# Data Directory

This directory contains runtime data files for the Blockhead Consulting website.

## Files

- **bookings.json** - Stores consultation booking data and time slot availability
  - Generated automatically when the server starts
  - Contains sensitive client information (emails, names, project details)
  - Should NOT be committed to git (excluded via .gitignore)
  - Automatically recreated if missing

## Privacy & Security

All files in this directory contain runtime data that may include:
- Client contact information
- Booking details and appointments
- Personal and business information

These files are excluded from version control for privacy and security reasons.

## Development

For local development:
1. The server will automatically create `bookings.json` on first run
2. The file will be populated with available time slots
3. Booking data persists between server restarts
4. Delete the file to reset all bookings and regenerate time slots