# Reminder App

This is a simple reminder application that allows users to create, manage, and organize reminders for family members. The application is structured to facilitate easy management of reminders and family members.

## Project Structure

```
reminder-app
├── cmd
│   └── main.go          # Entry point of the application
├── internal
│   ├── reminder
│   │   └── reminder.go  # Reminder struct and methods
│   └── family
│       └── family.go    # Family struct and methods
├── schemas
│   └── family_reminder.schema.json  # JSON schema for family reminders
├── go.mod               # Module definition and dependencies
└── README.md            # Project documentation
```

## Setup Instructions

1. **Clone the repository:**
   ```
   git clone <repository-url>
   cd reminder-app
   ```

2. **Install dependencies:**
   ```
   go mod tidy
   ```

3. **Run the application:**
   ```
   go run cmd/main.go
   ```

## Usage

- The application allows you to create reminders with a title, description, due date, and completion status.
- You can manage family members and associate reminders with them.
- The JSON schema located in `schemas/family_reminder.schema.json` defines the structure for family reminders.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue for any enhancements or bug fixes.