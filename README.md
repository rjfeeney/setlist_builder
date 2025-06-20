# **Setlist Builder CLI**

## Description:
This CLI is designed to take spotify playlists as input, capture and store metadata from songs in the playlist, and use this stored data to craft setlists for musical performances.

## Motivation:
As a performing musician myself, I have spent many hours crafting setlists by hand for various bars, weddings, and private events. Over the years, my bandmates and I developed rules for building setlists (what songs shouldn't go next to eachother, how to balance requests throughtout a setlist, etc.). Even still, making them by hand every time could take up to an hour per setlist, and reusing the same setlists from made us tired of those songs very quickly. I wanted to create a program that could follow the rules we had made for ourselves and reduce the time it took for us to make new setlists down to seconds.

I especially wanted the final product to be something that could be used by the average musician, whether they understood coding or not. To that end, I sought out to make a user-friendly setlist builder that could be used by the co-founder of my wedding band company (who has no coding experience whatsoever).

## 🚀 Quick Start

### Prerequisites
- Go 1.19 or higher
- Python 3.7 or higher (for API support)
- Docker & Docker Compose
- pip (Python package manager)

### Steps

```bash
# 1. Clone the repository
git clone https://github.com/rjfeeney/setlist_builder.git
cd setlist_builder

# 2. Create your environment variables file
cp .env.example .env
# Then edit .env and enter your Spotify API credentials

# 3. Start the PostgreSQL sample database
docker-compose up -d

# 4. Install Python dependencies
pip install -r requirements.txt

# 5. Run the CLI with sample data
go run main.go build --sample

# 6. Explore available commands
go run main.go --help
```

### 🎵 Spotify API Setup
To use the Spotify features, you’ll need your own developer credentials:

1. Visit https://developer.spotify.com/dashboard  
2. Create an app and copy your **Client ID** and **Client Secret**  
3. Paste them into your `.env` file like so:

```env
SPOTIFY_CLIENT_ID=your_client_id
SPOTIFY_CLIENT_SECRET=your_client_secret
```
