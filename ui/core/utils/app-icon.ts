/**
 * Maps common desktop app names → Simple Icons slugs
 * Full list at https://simpleicons.org
 */
const APP_ICON_MAP: Record<string, string> = {
    // Browsers
    "google chrome": "googlechrome",
    chrome: "googlechrome",
    firefox: "firefox",
    "mozilla firefox": "firefox",
    safari: "safari",
    edge: "microsoftedge",
    "microsoft edge": "microsoftedge",
    opera: "opera",
    brave: "brave",
    "brave browser": "brave",
    arc: "arc",

    // Editors / IDEs
    "visual studio code": "visualstudiocode",
    "vs code": "visualstudiocode",
    vscode: "visualstudiocode",
    "visual studio": "visualstudio",
    intellij: "intellijidea",
    "intellij idea": "intellijidea",
    webstorm: "webstorm",
    pycharm: "pycharm",
    goland: "goland",
    rider: "rider",
    clion: "clion",
    vim: "vim",
    neovim: "neovim",
    emacs: "gnuemacs",
    "sublime text": "sublimetext",
    "atom editor": "atom",
    cursor: "cursor",
    zed: "zedindustries",

    // Terminals
    terminal: "windowsterminal",
    "windows terminal": "windowsterminal",
    iterm2: "iterm2",
    warp: "warp",
    powershell: "powershell",
    "windows powershell": "powershell",
    cmd: "windowsterminal",
    bash: "gnubash",
    zsh: "zsh",
    hyper: "hyper",

    // Communication
    slack: "slack",
    discord: "discord",
    zoom: "zoom",
    teams: "microsoftteams",
    "microsoft teams": "microsoftteams",
    telegram: "telegram",
    whatsapp: "whatsapp",
    signal: "signal",
    skype: "skype",
    "google meet": "googlemeet",

    // Productivity
    notion: "notion",
    obsidian: "obsidian",
    figma: "figma",
    "adobe photoshop": "adobephotoshop",
    photoshop: "adobephotoshop",
    "adobe illustrator": "adobeillustrator",
    illustrator: "adobeillustrator",
    "adobe premiere": "adobepremierepro",
    "adobe after effects": "adobeaftereffects",
    xd: "adobexd",
    sketch: "sketch",

    // Dev tools
    docker: "docker",
    "docker desktop": "docker",
    postman: "postman",
    insomnia: "insomnia",
    github: "github",
    "github desktop": "github",
    gitkraken: "gitkraken",
    sourcetree: "sourcetree",
    tableplus: "tableplus",
    "db beaver": "dbeaver",

    // Music / Media
    spotify: "spotify",
    vlc: "vlc",
    "vlc media player": "vlc",

    // Misc
    finder: "apple",
    explorer: "windows",
    "windows explorer": "windows",
    notepad: "windows",
    word: "microsoftword",
    "microsoft word": "microsoftword",
    excel: "microsoftexcel",
    "microsoft excel": "microsoftexcel",
    powerpoint: "microsoftpowerpoint",
    "microsoft powerpoint": "microsoftpowerpoint",
    outlook: "microsoftoutlook",
    "microsoft outlook": "microsoftoutlook",
};

/**
 * Returns a Simple Icons CDN URL for known apps,
 * a Google favicon URL for URL-based entries,
 * or null to fall back to the letter avatar.
 */
export function getAppIconUrl(appName: string, url?: string): string | null {
    // URL-based: use Google favicon service
    if (url) {
        try {
            const domain = url.startsWith("http")
                ? new URL(url).hostname
                : url.split("/")[0];
            if (domain) {
                return `https://www.google.com/s2/favicons?domain=${domain}&sz=32`;
            }
        } catch {
            // ignore
        }
    }

    const slug = APP_ICON_MAP[appName.toLowerCase().trim()];
    if (slug) {
        return `https://cdn.simpleicons.org/${slug}/555555`;
    }

    return null;
}

/** Returns 1-2 letter initials from an app name */
export function getAppInitials(appName: string): string {
    const words = appName.trim().split(/\s+/);
    if (words.length === 1) return words[0].slice(0, 2).toUpperCase();
    return (words[0][0] + words[1][0]).toUpperCase();
}
