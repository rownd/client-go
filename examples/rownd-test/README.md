
```bash
git clone https://github.com/yourusername/rownd-test.git
cd rownd-test
```

2. Create a `.env` file in the project root:
```bash
ROWND_APP_KEY=your_app_key
ROWND_APP_SECRET=your_app_secret
ROWND_APP_ID=your_app_id
```

3. Get your Rownd credentials:
   - Log in to the [Rownd Dashboard](https://app.rownd.io)
   - Navigate to "Settings" > "API Keys"
   - Copy your App ID, App Key, and App Secret
   - Paste them into the `.env` file

4. Update the client-side configuration:
   - Open `client/static/index.html`
   - Find the Rownd initialization script
   - Replace the `ROWND_APP_KEY` value with your App Key

5. Start the server:
```bash
cd server
go run main.go
```

6. Open your browser to `http://localhost:8080`

## Features
- User authentication
- Token validation
- User profile management
- Group creation and management
- Group invitations

## Project Structure
- `/client` - Frontend HTML/JS/CSS
- `/server` - Go backend server
- `/pkg/rownd` - Rownd Go SDK implementation

## Reference Documentation
For more details on the Rownd API, visit:
https://docs.rownd.io/
```

