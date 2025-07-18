# Sample Chat Scenarios - Standard AI Agent

This document provides realistic chat examples demonstrating the AI agent's capabilities across different user roles and business scenarios.

## Table of Contents

1. [Guest User Scenarios](#guest-user-scenarios)
2. [Standard User Scenarios](#standard-user-scenarios)
3. [Manager Scenarios](#manager-scenarios)
4. [Developer Scenarios](#developer-scenarios)
5. [Admin Scenarios](#admin-scenarios)
6. [Policy Enforcement Examples](#policy-enforcement-examples)
7. [Business Workflow Examples](#business-workflow-examples)

---

## Guest User Scenarios

### Scenario 1: Basic Information Requests

```
🤖 Standard AI Agent with HGuard Integration
===============================================

✅ Logged in as: guest_user (guest)
🔐 Permissions:
📊 Session ID: session_1704067200

🛠️  Available Tools (3):
  - weather: Get current weather information for any city
  - addition: Perform mathematical addition of two numbers
  - search: Search the web for information

💬 Chat with the agent (type 'exit' to quit, 'help' for commands):
===============================================

You: What's the weather in London?

🤖 Agent: Weather in London: overcast clouds

You: Can you add 25 and 37?

🤖 Agent: The sum of 25.00 and 37.00 is 62.00

You: Search for the latest AI developments

🤖 Agent: Search results for: latest AI developments

This is a simulated search. In a real implementation, this would connect to a search API.

You: Can you send an email to my team?

🤖 Agent: Tool call rejected: Guests cannot send emails

You: Why can't I send emails?

🤖 Agent: As a guest user, you have limited access to protect system security. Email sending requires at least standard user privileges. You can access weather information, basic calculations, and web search. To send emails and access more tools, please register for a user account.
```

### Scenario 2: Guest Trying Administrative Functions

```
You: List all users in the system

🤖 Agent: Tool call rejected: Guests have limited access - please register

You: Show me system information

🤖 Agent: Tool call rejected: Guests have limited access - please register

You: What can I do as a guest?

🤖 Agent: As a guest user, you have access to these tools:
• **Weather**: Get current weather for any city
• **Addition**: Perform mathematical calculations
• **Search**: Search the web for information

For more advanced features like email, calendar, analytics, and system management, you'll need to register for a user account or higher privilege level.
```

---

## Standard User Scenarios

### Scenario 3: Daily Productivity Tasks

```
✅ Logged in as: standard_user (user)
🔐 Permissions: email_send
📊 Session ID: session_1704067300

🛠️  Available Tools (8):
  - weather: Get current weather information for any city
  - addition: Perform mathematical addition of two numbers
  - search: Search the web for information
  - send_email: Send email notifications (requires permission)
  - calendar: Manage calendar events and appointments
  - task_management: Create and manage tasks
  - document_gen: Generate documents from templates
  - notification: Send notifications to users

You: Create a task to review project documentation with high priority

🤖 Agent: Created task: review project documentation (Priority: high)

This is a simulated task creation. In a real implementation, this would create an actual task in your task management system.

You: Schedule a meeting for tomorrow at 2 PM with the development team

🤖 Agent: Created calendar event: meeting on tomorrow at 2 PM

This is a simulated calendar event creation.

You: Send an email to dev-team@company.com about the new project requirements

🤖 Agent: Email sent successfully!
To: dev-team@company.com
Subject: New Project Requirements
Body: Hello team, I wanted to inform you about the new project requirements that we need to discuss. Please review the latest documentation and let me know if you have any questions.

This is a simulated email. In a real implementation, this would send an actual email.

You: Generate a project report document

🤖 Agent: Generated report document using template: project_report
Data: map[date:2024-01-15 project:Current Project status:In Progress]

Document ID: DOC-1704067320
This is a simulated document generation.
```

### Scenario 4: User Hitting Permission Limits

```
You: Show me the sales analytics for this month

🤖 Agent: Tool call rejected: Analytics require elevated privileges

You: Can you query the user database?

🤖 Agent: Tool call rejected: Database access requires elevated privileges

You: What analytics can I access?

🤖 Agent: As a standard user, you don't have access to analytics tools. These features are available to managers, developers, and admins. However, you can:
• Create and manage your own tasks
• Schedule calendar events
• Send emails (with rate limits)
• Generate basic documents
• Access weather and search information

For analytics access, you would need manager or higher privileges.
```

---

## Manager Scenarios

### Scenario 5: Business Operations Management

```
✅ Logged in as: manager_user (manager)
🔐 Permissions: financial_access, email_send, database_query
📊 Session ID: session_1704067400

🛠️  Available Tools (12):
  - weather: Get current weather information for any city
  - addition: Perform mathematical addition of two numbers
  - search: Search the web for information
  - quote: Get financial quotes for currency exchanges
  - price: Get current currency pricing information
  - send_email: Send email notifications (requires permission)
  - database_query: Query databases (requires permission)
  - calendar: Manage calendar events and appointments
  - task_management: Create and manage tasks
  - analytics: Generate analytics reports and insights
  - document_gen: Generate documents from templates
  - notification: Send notifications to users

You: Generate a sales report for this month

🤖 Agent: Sales Report (month):
- Total Revenue: $125,000
- Orders: 1,250
- Average Order Value: $100
- Growth: +15% vs previous period

This is simulated analytics data.

You: Get a quote for converting 50,000 USD to EUR

🤖 Agent: Quote service unavailable - API key not configured

You: Query the database for active customers

🤖 Agent: Database Query Results:
Database: customer_db
Query: SELECT * FROM customers WHERE status = 'active'

Results:
- Record 1: Sample customer data
- Record 2: More sample customer data

This is a simulated database query. In a real implementation, this would query an actual database.

You: Create a high-priority task for the team to review Q4 results

🤖 Agent: Created task: review Q4 results (Priority: high)

This is a simulated task creation. In a real implementation, this would create an actual task in your task management system.

You: Send SMS notifications to all team leads about the quarterly review

🤖 Agent: Sent sms notification to 3 recipients
Message: Quarterly review meeting scheduled for Friday at 10 AM. Please prepare your department reports.

This is a simulated notification.
```

### Scenario 6: Manager Financial Operations

```
You: Check Bitcoin price

🤖 Agent: Tool call rejected: Price queries only available during business hours (9-17)

You: What time is it?

🤖 Agent: I can see from your session that it's currently outside business hours. Price queries are restricted to business hours (9 AM - 5 PM) for security reasons. You can try again during business hours, or contact an admin for after-hours access.

You: Generate a user analytics report

🤖 Agent: User Analytics (month):
- Active Users: 5,420
- New Signups: 234
- Retention Rate: 87%
- Churn Rate: 3.2%

This is simulated analytics data.

You: Delete the old marketing task

🤖 Agent: Deleted task: old marketing task

This is a simulated task deletion.
```

---

## Developer Scenarios

### Scenario 7: Technical Operations

```
✅ Logged in as: developer_user (developer)
🔐 Permissions: system_access, file_operations, database_query
📊 Session ID: session_1704067500

🛠️  Available Tools (11):
  - weather: Get current weather information for any city
  - addition: Perform mathematical addition of two numbers
  - search: Search the web for information
  - file_operations: Manage files and directories (admin only)
  - system_info: Get system information and metrics
  - send_email: Send email notifications (requires permission)
  - database_query: Query databases (requires permission)
  - calendar: Manage calendar events and appointments
  - task_management: Create and manage tasks
  - analytics: Generate analytics reports and insights
  - document_gen: Generate documents from templates

You: Show me system CPU information

🤖 Agent: CPU Usage: 45%
CPU Model: Intel Core i7-9700K
Cores: 8

You: List files in the /var/log directory

🤖 Agent: Listing files in directory: /var/log

This is a simulated file listing. In a real implementation, this would list actual files.

You: Read the application.log file

🤖 Agent: Reading file: application.log

This is a simulated file read. In a real implementation, this would read the actual file content.

You: Query the database for recent error logs

🤖 Agent: Database Query Results:
Database: logs
Query: SELECT * FROM error_logs WHERE timestamp > '2024-01-14'

Results:
- Record 1: Sample error log entry
- Record 2: More sample error log data

This is a simulated database query. In a real implementation, this would query an actual database.

You: Check system memory usage

🤖 Agent: Memory Usage: 8.2GB / 16GB (51%)
Available: 7.8GB

You: Generate a performance report

🤖 Agent: Performance Report (month):
- Average Response Time: 125ms
- Error Rate: 0.02%
- Uptime: 99.98%
- Throughput: 1,250 req/sec

This is simulated analytics data.
```

### Scenario 8: Developer File Operations

```
You: Delete the old config file

🤖 Agent: Tool call rejected: File deletion requires admin privileges

You: Write to a new configuration file

🤖 Agent: Writing to file: config.json
Content: {"database": "localhost", "port": 3306, "ssl": true}

This is a simulated file write. In a real implementation, this would write to the actual file.

You: Check network status

🤖 Agent: Network Status: Connected
Bandwidth: 100 Mbps
Latency: 12ms

You: What's the difference between my permissions and admin permissions?

🤖 Agent: As a developer, you have technical access including:
• System information and monitoring
• File operations (read, write, list - but not delete)
• Database queries
• Performance analytics

Admin permissions additionally include:
• File deletion capabilities
• User management
• System administration
• Override most security restrictions
• After-hours access to restricted tools

You have strong technical access but admins have full system control.
```

---

## Admin Scenarios

### Scenario 9: System Administration

```
✅ Logged in as: admin_user (admin)
🔐 Permissions: financial_access, system_access, user_management, file_operations, email_send, database_query
📊 Session ID: session_1704067600

🛠️  Available Tools (15):
  - weather: Get current weather information for any city
  - addition: Perform mathematical addition of two numbers
  - search: Search the web for information
  - quote: Get financial quotes for currency exchanges
  - price: Get current currency pricing information
  - file_operations: Manage files and directories (admin only)
  - system_info: Get system information and metrics
  - user_management: Manage user accounts and roles (admin only)
  - send_email: Send email notifications (requires permission)
  - database_query: Query databases (requires permission)
  - calendar: Manage calendar events and appointments
  - task_management: Create and manage tasks
  - analytics: Generate analytics reports and insights
  - document_gen: Generate documents from templates
  - notification: Send notifications to users

You: List all users in the system

🤖 Agent: Active Users:
- admin (Admin)
- john_doe (Manager)
- jane_smith (User)
- dev_user (Developer)

This is a simulated user listing.

You: Create a new user account for sarah_jones with manager role

🤖 Agent: Created user: sarah_jones with role: manager

This is a simulated user creation. In a real implementation, this would create an actual user account.

You: Delete the old backup file

🤖 Agent: Deleting file: backup_old.sql

This is a simulated file deletion. In a real implementation, this would delete the actual file.

You: Update john_doe's role to admin

🤖 Agent: Updated user: john_doe to role: admin

This is a simulated user update.

You: Get a large currency quote for 100,000 USD

🤖 Agent: Quote service unavailable - API key not configured

You: Generate a comprehensive system performance report

🤖 Agent: Performance Report (month):
- Average Response Time: 125ms
- Error Rate: 0.02%
- Uptime: 99.98%
- Throughput: 1,250 req/sec

This is simulated analytics data.
```

### Scenario 10: Admin Override Capabilities

```
You: Check Bitcoin price

🤖 Agent: Price service unavailable - API key not configured

You: Query the database for all user login attempts

🤖 Agent: Database Query Results:
Database: security
Query: SELECT * FROM login_attempts WHERE date > '2024-01-01'

Results:
- Record 1: Sample login data
- Record 2: More sample login data

This is a simulated database query. In a real implementation, this would query an actual database.

You: Send emergency notification to all users

🤖 Agent: Sent email notification to 5 recipients
Message: Emergency system maintenance will begin in 30 minutes. Please save your work and log out.

This is a simulated notification.

You: Delete user account for inactive_user

🤖 Agent: Deleted user: inactive_user

This is a simulated user deletion. In a real implementation, this would delete the actual user account.
```

---

## Policy Enforcement Examples

### Scenario 11: Time-Based Restrictions

```
You: Query the customer database

🤖 Agent: Tool call rejected: Database access only available during business hours

You: But I'm an admin, doesn't that override time restrictions?

🤖 Agent: Even as an admin, certain security policies apply during off-hours. Database access is restricted to business hours (8 AM - 6 PM) for security reasons. If you need emergency access, you would need to use the emergency override procedures or contact the security team.

You: What can I do right now?

🤖 Agent: During off-hours, you can still:
• Access weather and search information
• Perform calculations
• Use file operations
• Check system information
• Send emails and notifications
• Manage calendar and tasks
• Generate documents

Database queries and some financial operations are restricted until business hours resume.
```

### Scenario 12: Rate Limiting in Action

```
You: Send an email to team@company.com

🤖 Agent: Email sent successfully!
To: team@company.com
Subject: Update
Body: Please check the latest project status

[... after several email attempts ...]

You: Send another email to support@company.com

🤖 Agent: Tool call rejected: Email rate limit exceeded for this session

You: How many emails have I sent?

🤖 Agent: You've reached the session limit for email sending. This prevents spam and ensures responsible usage. The limit resets with a new session, or you can wait for the current session to cool down. As an alternative, you can use the notification system for urgent messages.
```

### Scenario 13: Parameter-Based Restrictions

```
You: Get a quote for 50,000 USD to EUR transaction

🤖 Agent: Tool call rejected: Large transactions require admin approval

You: But I am an admin!

🤖 Agent: You're right! Let me process that large transaction for you.

🤖 Agent: Quote service unavailable - API key not configured

You: Get a quote for 5,000 USD to EUR

🤖 Agent: Quote service unavailable - API key not configured
```

---

## Business Workflow Examples

### Scenario 14: Project Management Workflow

```
You: Create a new project task for Q1 planning

🤖 Agent: Created task: Q1 planning (Priority: medium)

You: Schedule a planning meeting for next Monday at 10 AM

🤖 Agent: Created calendar event: planning meeting on next Monday at 10 AM

You: Send an email to project-team@company.com about the planning session

🤖 Agent: Email sent successfully!
To: project-team@company.com
Subject: Q1 Planning Session
Body: Hi team, I've scheduled our Q1 planning session for next Monday at 10 AM. Please come prepared with your department goals and resource requirements.

You: Generate a project planning document template

🤖 Agent: Generated presentation document using template: project_planning
Data: map[quarter:Q1 year:2024 meeting_date:next Monday]

Document ID: DOC-1704067800
This is a simulated document generation.

You: Update the task status to in progress

🤖 Agent: Updated task Q1-planning status to: in_progress

This is a simulated task update.
```

### Scenario 15: Customer Support Workflow

```
You: Search for recent customer complaints

🤖 Agent: Search results for: recent customer complaints

This is a simulated search. In a real implementation, this would connect to a search API.

You: Query the database for support tickets from last week

🤖 Agent: Database Query Results:
Database: support
Query: SELECT * FROM tickets WHERE created_date > '2024-01-08'

Results:
- Record 1: Sample support ticket
- Record 2: More sample ticket data

This is a simulated database query. In a real implementation, this would query an actual database.

You: Generate a customer service report

🤖 Agent: Analytics Report (week):
- Overall system health: Excellent
- Key metrics trending positive

This is simulated analytics data.

You: Send notification to support team about high priority tickets

🤖 Agent: Sent email notification to 1 recipients
Message: We have several high priority support tickets that need immediate attention. Please check your queue.

This is a simulated notification.
```

### Scenario 16: Financial Operations Workflow

```
You: Check current Bitcoin price

🤖 Agent: Price service unavailable - API key not configured

You: Generate a sales analytics report for the quarter

🤖 Agent: Sales Report (quarter):
- Total Revenue: $125,000
- Orders: 1,250
- Average Order Value: $100
- Growth: +15% vs previous period

This is simulated analytics data.

You: Get a quote for a 25,000 USD international transfer

🤖 Agent: Quote service unavailable - API key not configured

You: Generate a financial summary document

🤖 Agent: Generated report document using template: financial_summary
Data: map[period:Q1 2024 revenue:$125,000 growth:+15%]

Document ID: DOC-1704067900
This is a simulated document generation.

You: Send the financial report to finance-team@company.com

🤖 Agent: Email sent successfully!
To: finance-team@company.com
Subject: Q1 Financial Summary
Body: Please find attached the Q1 financial summary report showing our revenue growth and key metrics.

This is a simulated email. In a real implementation, this would send an actual email.
```

---

## Advanced Scenarios

### Scenario 17: Multi-Step Complex Task

```
You: I need to prepare for our quarterly board meeting

🤖 Agent: I'll help you prepare for the quarterly board meeting. This typically involves several steps. Let me start by gathering the necessary information and documents.

You: Generate a comprehensive quarterly report

🤖 Agent: Generated report document using template: quarterly_report
Data: map[quarter:Q4 year:2023 revenue:$500,000 growth:+12%]

Document ID: DOC-1704068000
This is a simulated document generation.

You: Get sales analytics for the quarter

🤖 Agent: Sales Report (quarter):
- Total Revenue: $125,000
- Orders: 1,250
- Average Order Value: $100
- Growth: +15% vs previous period

This is simulated analytics data.

You: Schedule the board meeting for next Friday at 2 PM

🤖 Agent: Created calendar event: board meeting on next Friday at 2 PM

You: Send meeting invitations to board@company.com

🤖 Agent: Email sent successfully!
To: board@company.com
Subject: Quarterly Board Meeting - January 19, 2024
Body: You're invited to our quarterly board meeting on Friday, January 19th at 2 PM. We'll review Q4 results and discuss Q1 strategy.

You: Create a task to prepare presentation materials

🤖 Agent: Created task: prepare presentation materials (Priority: high)

You: Perfect! Is there anything else I should prepare?

🤖 Agent: For a comprehensive board meeting preparation, you might also consider:

• Reviewing competitor analysis (use search tool)
• Preparing financial projections (analytics tool)
• Creating action items for Q1 (task management)
• Setting up follow-up notifications for attendees
• Generating supporting documents like budget reports

Would you like me to help with any of these additional items?
```

### Scenario 18: Error Handling and Recovery

```
You: managment list users

🤖 Agent: Active Users:
- admin (Admin)
- john_doe (Manager)
- jane_smith (User)
- dev_user (Developer)

This is a simulated user listing.

You: wheather in Paris

🤖 Agent: Weather in Paris: partly cloudy

You: Can you fix that spelling correction?

🤖 Agent: Yes! The system automatically corrected "wheather" to "weather" and "managment" to "user_management" using built-in auto-correction policies. This helps prevent common typos from causing errors.

The auto-correction features include:
• "wheather" → "weather"
• "calander" → "calendar"
• "managment" → "user_management"

This makes the system more user-friendly while maintaining accuracy.
```

---

## Usage Notes

1. **API Keys**: Most external services show "service unavailable" messages since API keys aren't configured. In production, these would connect to real services.

2. **Simulated Responses**: Tools show "This is a simulated..." messages to indicate they're demonstration implementations. Real implementations would perform actual operations.

3. **Role-Based Access**: Each role has different available tools and permissions, demonstrating the security model.

4. **Policy Enforcement**: The system actively enforces time-based, parameter-based, and role-based restrictions.

5. **Conversation Flow**: The agent maintains context and can handle both tool calls and natural conversation.

6. **Error Handling**: The system provides clear error messages and suggests alternatives when access is denied.

These examples show how the standard AI agent can be used across different business scenarios while maintaining security and providing rich functionality appropriate to each user's role and context.
