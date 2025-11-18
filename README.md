# JobFlow  
Your AI-powered personal job tracking assistant.  
Scrape LinkedIn → Enrich with AI → Save to Notion with one click.

---

## 🚀 Overview

JobFlow is a personal automation tool that removes the most annoying part of job hunting:  
copy/pasting job descriptions into a spreadsheet.

With one click on a LinkedIn job posting:

- A Chrome extension scrapes job data  
- A local Go REST API captures and enriches the job  
- Everything is pushed into your Notion Job Tracker automatically

This is your private, local, AI-assisted ATS.

---

## ✨ Features

**Smart LinkedIn scraper**  
Automatically extracts:
- Job title  
- Company  
- Location  
- Work mode (Remote / Hybrid / On-site)  
- Full job description  
- Job URL  

Includes multiple fallback selectors and description-based inference for tricky pages.

**AI enrichment (optional)**  
- Summary  
- Key skills  
- Tailored notes  

**Notion integration**  
- Automatically creates new rows  
- Supports rich text, URLs, select fields, and dates  

**SQLite storage**  
- Tracks duplicates  
- Stores application stages  

**Modular Go architecture**  
- Clean separation between jobs, Notion helpers, and AI logic  

---

## 🧰 Requirements

- Go 1.22+  
- SQLite3  
- Chrome or Brave  
- Notion integration token  
- A Notion database with matching properties  

---

## 📦 Project Structure

```
jobflow/
├── cmd/jobflow/             # Main server
├── internal/
│   ├── jobs/                # SQLite logic
│   ├── notion/              # Notion API wrapper
│   ├── ai/                  # LLM enrichment (optional)
├── chrome-extension/
│   ├── popup.js             # Scraper + send-to-API
│   ├── popup.html
│   ├── manifest.json
```

---

## 🔧 Setup

### 1. Clone the repository

```
git clone https://github.com/YOUR_USERNAME/jobflow
cd jobflow
```

### 2. Add environment variables

Create a `.env` file:

```
NOTION_TOKEN=your_notion_token
NOTION_DATABASE_ID=your_database_id
OPENAI_API_KEY=optional_openai_key
```

### 3. Run the server

```
go run ./cmd/jobflow
```

Expected output:

```
Notion connection OK.
HTTP listening on :8081
```

### 4. Install the Chrome extension

1. Go to `chrome://extensions`
2. Enable Developer Mode
3. Click "Load Unpacked"
4. Select the `chrome-extension/` folder

---

## 🖱️ Usage

1. Open any LinkedIn job posting  
2. Click the JobFlow button  
3. A new row appears in your Notion Job Tracker instantly  

---

## 🧪 Testing

You can test without LinkedIn using:

```
curl -X POST http://localhost:8081/apply \
  -H "Content-Type: application/json" \
  -d '{"position":"Test Role","company":"TestCorp"}'
```

---

## 🔮 Roadmap

- Salary inference  
- Recruiter extraction  
- Auto-drafted outreach messages  
- Application analytics and dashboards  
- Cloud deployment  
- Chrome Web Store release  

---

## 🧑‍💻 Author

**Yassine Berrada**  
AI Engineer & Software Developer  
Building tools that give humans superpowers.
