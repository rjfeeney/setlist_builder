# **Setlist Builder CLI**

## Description:
This CLI is designed to take spotify playlists as input, capture and store metadata from songs in the playlist, and use this stored data to craft setlists for musical performances.

## Motivation:
As a performing musician myself, I have spent many hours crafting setlists by hand for various bars, weddings, and private events. Over the years, my bandmates and I developed rules for building setlists (what songs shouldn't go next to eachother, how to balance requests throughtout a setlist, etc.). Even still, making them by hand every time could take up to an hour per setlist, and reusing the same setlists from previous shows made us tired of those songs very quickly. I wanted to create a program that could follow the rules we had made for ourselves and reduce the time it took for us to make new setlists down to seconds.

I especially wanted the final product to be something that could be used by the average musician, whether they understood coding or not. To that end, I sought out to make a user-friendly setlist builder that could be used by the co-founder of my wedding band company (who has no coding experience whatsoever). While a more layman's version is still in development, the CLI version is in working condition and I use it regularly to automate this lengthy task.

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Git

### Steps

## 1. Clone the repository
```bash
git clone https://github.com/rjfeeney/setlist_builder.git
cd setlist_builder
```

## 2. Create your environment variables file
```bash
cp .env.sample .env
```

## 3. Obtain Spotify for Devs Credentials
To use the Spotify features, you’ll need your own developer credentials:

1. Visit https://developer.spotify.com/dashboard  
2. Create an app and copy your **Client ID** and **Client Secret**  
3. Paste them, as well as the database URL into your `.env` file like so:

```bash
SPOTIFY_CLIENT_ID=your_client_id
SPOTIFY_CLIENT_SECRET=your_client_secret
DATABASE_URL="postgres://postgres:postgres@localhost:5432/setlist-builder?sslmode=disable"
```

## 4.  Start the PostgreSQL sample database
```bash
docker-compose up -d
```

## 5. Build the CLI binary (Cross-Platform)
Run the appropriate build script for your OS:
* On **macOS/Linux:**
```bash
chmod +x build.sh
./build.sh
```
* On **Windows (Powershell):**
* ```bash
.\build.ps1
```
This will build the binary inside Docker and place it in the ./output directory. It is recommended that you operate from within the output directory to reduce the amount of typing needed to execute commands:
```bash
cd output
```

## 6. Run the CLI with sample data
```bash
./setlist build    # macOS/Linux
.\setlist build    # Windows
```

## 7. Explore available commands
```bash
./setlist help    # macOS/Linux
.\setlist help    # Windows
```

## Usage

### Command Format
All commands must begin with `./setlist` (macOS/Linux) or `.\setlist` (Windows followed by the CLI command:

```bash
./setlist [command] [required-parameter] {optional-parameter}    # macOS/Linux
.\setlist [command] [required-parameter] {optional-parameter}    # Windows
```

### Command List

**Extract [Spotify Playlist URL]**
- Extracts metadata from all tracks in a Spotify playlist and stores it in the database. Songs already in the database will be skipped over. If songs fail to be added for whatever reason, try running the clean function (detailed below) and then rerunning the extract command. Note that analysis of metadata is not guaranteed to be 100% accurate.

**List**
- Lists all songs currently in the tracks table in the database.

**Singers**
- Searches through all tracks in the singers table and prompts you to enter singer and key info for any songs with no singer listed.
- Future improvements will add in the feature to access songs that already have singers and keys.

**Keys {missing}**
- Searches for tracks in the tracks table and prompts you to enter original key info.
- Including 'missing' in the command will iterate through all tracks in the tracks table with missing original keys.
- Otherwise, you will be prompted to enter a song title to look up. Note that spelling must be exact (but it is not case sensitive).

**Clean [table]**
- Removes any tracks from the database that are missing info (usually key and bpm).
- Use this before rerunning the extract command for any tracks that didn't make it on the first try.

**Build**
- Begins the setlist building process, starting with a questions about set length, requests, and 'Do Not Plays'.
- Upon successful creation, the setlist will print to your terminal.

**Database**
- Allows for manual access to the database to make changes as needed.
- This is only advised to those who are comfortable writing SQL commands.
- You will be asked to confirm direct database access before connecting.

**Clear [table]**
- Clears specified table from the database.
- Use this if you need to reset singers, tracks, or the working table.

**Reset**
- Clears the entire database of all tracks.
- Note that you should only do this if the data has somehow become corrupted or unuseable, as extracting new songs is the lengthiest part of the process.

**Help**
- Lists all available commands and their descriptions.


## Contributing

### Clone the repo
```bash
git clone https://github.com/rjfeeney/setlist_builder
cd setlist_builder
```

### Create your environment variables file
```bash
cp .env.example .env
```

### Obtain Spotify for Devs Credentials
To use the Spotify features, you’ll need your own developer credentials:

1. Visit https://developer.spotify.com/dashboard  
2. Create an app and copy your **Client ID** and **Client Secret**  
3. Paste them, as well as the database URL into your `.env` file like so:

```bash
SPOTIFY_CLIENT_ID=your_client_id
SPOTIFY_CLIENT_SECRET=your_client_secret
DATABASE_URL="postgres://postgres:postgres@localhost:5432/setlist-builder?sslmode=disable"
```

### Build the project
```bash
docker-compose up -d
```

### Making changes
1. Create a feature branch from `main`
2. Make your changes
3. Test locally to ensure everything works
4. Commit with a clear, descriptive message
5. Push and open a pull request

### Testing changes
This project currently uses manual testing. After making changes, test your modifications by running the application locally and verifying the functionality works as expected.

### Submit a pull request
If you'd like to contribute, please fork the repository and open a pull request to the `main` branch.

