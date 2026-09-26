package main

// Everything a visitor can read, in one place. The curl résumé, the SSH app
// and the no-terminal fallback all render from these values, so a change here
// shows up everywhere at once.
//
// Two lists on purpose. resumeWork mirrors the one page B/W résumé (the source
// of truth for the résumé). portfolio mirrors the work list on the homepage.
// Keep each one in step with the page it copies, and keep the numbers honest.

const (
	name     = "Mosam Biswas"
	role     = "Software engineer · AI / ML · Full-stack"
	degree   = "B.Tech AI & DS · Class of 2026"
	place    = "Navi Mumbai, India"
	status   = "Open to full-time roles & internships"
	email    = "mosambiswas999@gmail.com"
	site     = "https://www.mosambiswas.com"
	github   = "https://github.com/BiswasMosam"
	linkedin = "https://linkedin.com/in/mosambiswas"
	pdf      = "https://www.mosambiswas.com/MosamBiswasResume.pdf"
)

const summary = "A 2026 B.Tech AI & Data Science graduate who builds AI systems and ships " +
	"them as working software: an AI form-filling Chrome extension, a fully local voice " +
	"assistant, a mental health platform used by 100+ students, and offline apps for two " +
	"small businesses."

type job struct {
	Role, Company, When, Where string
	Points                     []string
}

var experience = []job{{
	Role:    "Software Developer Intern",
	Company: "Sedna Technologies",
	When:    "Jan / Apr 2026",
	Where:   "Mumbai",
	Points: []string{
		"Shipped work across 4 client projects, working directly with clients to gather requirements and resolve issues.",
		"Designed and built a client-facing dashboard solo with Angular, Flask and MySQL over REST APIs, owning it from requirements to handover (later submitted as my B.Tech dissertation).",
	},
}}

type project struct {
	Name    string
	Tagline string // one line, what it is
	Year    string
	Stack   string
	Stat    string // the big number on the homepage preview card
	StatFor string // what the number means
	About   string // a short paragraph for the detail view
	URL     string
	Badge   string // "Live" or "Install" when there is something to use, not just read
}

var resumeWork = []project{
	{Name: "Fill.ai", Year: "2026", Stack: "JS · Chrome MV3 · LLMs", URL: site + "/Fill.ai/",
		About: "Chrome extension that fills web forms with AI from your resume. Every answer must cite a profile field or it is thrown out; 92-check end-to-end suite, released by CI/CD."},
	{Name: "AMINAL", Year: "2026", Stack: "Python · Whisper · Ollama", URL: github + "/Aminal-Public",
		About: "Local-first voice assistant for Windows: speech, LLM intent, vision and text-to-speech fully on-device. 29K lines of Python, 125 commands, 849 tests."},
	{Name: "MindFirst", Year: "2025", Stack: "Next.js · Flask · Prisma", URL: github + "/Mind-First",
		About: "Mental health pre-screening for 100+ students, chatbot + PHQ-9/GAD-7 at 95% accuracy, drop-off cut 30%."},
	{Name: "IntelliStatement", Year: "2025", Stack: "FastAPI · Streamlit · LLM", URL: github + "/LLM-bank-statement-analyzer",
		About: "Bank statement analyzer, LLM extraction from PDFs and images into structured data, insights via FastAPI + Streamlit."},
	{Name: "PixelForge", Year: "2025", Stack: "Python · PyTorch · GANs", URL: github + "/PixelForge",
		About: "Leukemia cell classification on C-NMC at 0.90 ROC-AUC, 95% recall; a custom spectral-norm GAN augments scarce data."},
	{Name: "Pops & Order Control", Year: "2026", Stack: "Flutter · Dart · SQLite",
		About: "Two offline Android apps for real small businesses: a fabrication firm's order tracker and a food stall order board, on relational SQLite."},
}

var portfolio = []project{
	{Name: "AMINAL", Tagline: "Local-first voice assistant for Windows", Year: "2026",
		Stack: "Python · Whisper · Ollama", Stat: "Local", StatFor: "voice, vision and memory on your own GPU, nothing leaves the machine",
		About: "Speech, LLM intent, vision and text-to-speech, all on-device. 29K lines of Python, 125 commands, 849 tests. An active project, still being extended.",
		URL:   github + "/Aminal-Public"},
	{Name: "Fill.ai", Tagline: "AI form-filling Chrome extension", Year: "2026", Badge: "Install",
		Stack: "JS · Chrome MV3 · LLMs", Stat: "92", StatFor: "end-to-end checks, every answer must cite your profile or it is thrown out",
		About: "Fills web forms from your resume. It learns from what you type, handles Google Forms and Select2, and refuses to invent: an answer with no profile field behind it never reaches the page. Released by CI/CD.",
		URL:   site + "/Fill.ai/"},
	{Name: "Sim Sim", Tagline: "Voice-driven AI desktop assistant", Year: "2024",
		Stack: "Python · PySide6", Stat: "70%", StatFor: "faster task execution, voice commands plus a conversational GUI",
		About: "An AI desktop assistant you can talk to or type at: voice commands for the quick things, a conversational GUI for the rest.",
		URL:   github + "/SimSim-Virtual-Assistant"},
	{Name: "Rail Mitra", Tagline: "Railway platform & track optimizer", Year: "2025",
		Stack: "Python · OR-Tools", Stat: "Zero", StatFor: "conflict schedules, constraint programming assigns platforms & tracks",
		About: "Assigns trains to platforms and tracks with constraint programming, so the schedule it hands back has no conflicts in it.",
		URL:   github + "/RailMind"},
	{Name: "MindFirst", Tagline: "Mental-health screening platform", Year: "2025",
		Stack: "Next.js · Flask · Prisma", Stat: "95%", StatFor: "screening accuracy, chatbot + psychometrics for 100+ students",
		About: "Mental health pre-screening for 100+ students: a chatbot plus PHQ-9 and GAD-7 at 95% accuracy, and 30% fewer people dropping off halfway.",
		URL:   github + "/Mind-First"},
	{Name: "PixelForge", Tagline: "Medical imaging with CNNs & GANs", Year: "2025",
		Stack: "PyTorch · GANs", Stat: "0.90", StatFor: "ROC-AUC on leukemia detection, custom GAN augments scarce data",
		About: "Leukemia cell classification on C-NMC at 0.90 ROC-AUC and 95% recall. Labelled cells are scarce, so a custom spectral-norm GAN makes more.",
		URL:   github + "/PixelForge"},
	{Name: "IntelliStatement", Tagline: "LLM bank-statement analyzer", Year: "2025",
		Stack: "FastAPI · Streamlit", Stat: "LLM", StatFor: "extraction from PDFs & images, insights via FastAPI + Streamlit",
		About: "Reads bank statements from PDFs and photos, pulls every transaction into structured data with an LLM, then charts where the money went.",
		URL:   github + "/LLM-bank-statement-analyzer"},
	{Name: "Committee Connect", Tagline: "Android app for college committees", Year: "2024",
		Stack: "Android · Java · Firebase", Stat: "Team", StatFor: "a four-person RAIT mini project, synced over Firebase",
		About: "A native Android app for running college committees, built by a team of four as a RAIT mini project, with Firebase Realtime Database for sync.",
		URL:   github + "/CommitteesManagement"},
	{Name: "PixelShift", Tagline: "In-browser batch image converter", Year: "2025", Badge: "Live",
		Stack: "Vanilla JS · Canvas", Stat: "Zero", StatFor: "uploads, batch image conversion entirely in the browser",
		About: "Drop in a batch of images and convert them all without anything leaving your machine. Runs entirely in the browser.",
		URL:   site + "/PixelShift/"},
	{Name: "Flex Card", Tagline: "Hardware-spec card generator", Year: "2024",
		Stack: "Python · PyQt5", Stat: "PNG", StatFor: "share-ready spec cards, scans your hardware, exports one click later",
		About: "A Windows desktop app that scans your hardware and exports a polished PNG summary card, ready to share.",
		URL:   github + "/Flex-Card"},
	{Name: "To-Do", Tagline: "Full-stack task boards with auth", Year: "2024", Badge: "Live",
		Stack: "Node.js · JS", Stat: "E2E", StatFor: "task boards with auth, Node backend to front-end, the whole loop",
		About: "Task boards with sign-in, from the Node backend to the front end. Small, but the whole loop.",
		URL:   site + "/To-Do/"},
}

type paper struct {
	Title, Venue, Role, Summary, URL string
}

var research = paper{
	Title:   "Comparative Analysis of Psychometric Data for Assessing Mental Health Testing Needs Using Machine Learning Models",
	Venue:   "IEEE Xplore · Proc. ICAIIHI 2025",
	Role:    "Co-author",
	Summary: "Benchmarked Logistic Regression, Random Forest, SVM, XGBoost and LightGBM on 28.5K PHQ-9 / GAD-7 records (stratified 5-fold CV, paired t-tests). Logistic Regression led at 99% accuracy with SHAP explanations, and the pre-screening flow roughly halves screening overhead.",
	URL:     site + "/ResearchPaper.pdf",
}

type school struct {
	Name, When, Detail string
}

var education = school{
	Name:   "Ramrao Adik Institute of Technology",
	When:   "2022 / 2026",
	Detail: "B.Tech in Computer Engineering (Major: Artificial Intelligence and Data Science) · D. Y. Patil Deemed to be University · CGPA 7.5 · Navi Mumbai",
}

type skill struct {
	Group, Items string
}

var stack = []skill{
	{"Languages", "Python, Java, JavaScript, TypeScript, Dart, SQL, C++, C, HTML/CSS"},
	{"AI / ML", "PyTorch, TensorFlow, scikit-learn, XGBoost, LightGBM, SHAP, OpenCV, LLMs (OpenAI, Gemini, Claude, Ollama)"},
	{"Backend & Web", "FastAPI, Flask, Node.js, Express.js, Next.js, React, Angular, REST APIs, WebSockets, Chrome Extensions"},
	{"Mobile & DB", "Flutter, Android (Java), Firebase, MySQL, PostgreSQL, MongoDB, SQLite"},
	{"Tools", "Git, GitHub Actions (CI/CD), AWS, Pandas, Postman, Google OR-Tools, Unit & E2E Testing, Agile/Scrum"},
}

var leadership = []string{
	"Technical Head · Photocircle RAIT, 2024 / 2025. Ran 3 workshops for 75+ students; participation up 60%.",
	"Google Developer Groups · 25+ codelabs completed, Google Cloud Innovator.",
	"Hacktoberfest, and the Smart India Hackathon 2025 internal round.",
}

// The About tab. Written the way the homepage talks.
var aboutText = []string{
	"I'm a B.Tech AI & Data Science graduate from RAIT, D.Y. Patil University, and I spent those years turning coursework into shipped software: ML pipelines, Android apps, web tools and one peer-reviewed paper.",
	"The honest version: I learn by building, break things politely, then fix them properly. I care about the whole loop, from a model's confusion matrix to the button a person actually presses. The messy middle, data cleaning, edge cases, the last 10 percent, is where most of the real work lives.",
	"Away from the keyboard I photograph things. I was Technical Head of Photocircle RAIT, and the pictures live at mosambiswas.com/sheichobi.",
}

var numbers = []struct{ N, What string }{
	{"20+", "projects shipped"},
	{"25+", "Google codelabs"},
	{"01", "IEEE publication"},
	{"03", "years behind the lens"},
}
