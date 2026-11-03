pragma Singleton
pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import Quickshell.Io
import "GreetdEnv.js" as GreetdEnv
import qs.Services

Singleton {
    id: root
    readonly property var log: Log.scoped("GreetdMemory")

    readonly property string greetCfgDir: Quickshell.env("DMS_GREET_CFG_DIR") || "/var/cache/dms-greeter"
    readonly property string memoryFile: greetCfgDir + "/.local/state/memory.json"
    readonly property bool rememberLastSession: GreetdEnv.readBoolOverride(Quickshell.env, ["DMS_GREET_REMEMBER_LAST_SESSION", "DMS_SAVE_SESSION"], true, log)
    readonly property bool rememberLastUser: GreetdEnv.readBoolOverride(Quickshell.env, ["DMS_GREET_REMEMBER_LAST_USER", "DMS_SAVE_USERNAME"], true, log)

    property string lastSessionId: ""
    property string lastSessionDesktopId: ""
    property string lastSuccessfulUser: ""
    property var userSessions: ({})
    property bool memoryReady: false

    Component.onCompleted: {
        parseMemory(memoryFileView.text());
    }

    function parseMemory(content) {
        try {
            if (!content || !content.trim())
                return;
            const memory = JSON.parse(content);
            lastSessionId = rememberLastSession ? (memory.lastSessionId || "") : "";
            lastSessionDesktopId = rememberLastSession ? (memory.lastSessionDesktopId || "") : "";
            lastSuccessfulUser = rememberLastUser ? (memory.lastSuccessfulUser || "") : "";
            userSessions = rememberLastSession ? parseUserSessions(memory.userSessions) : {};
            if (!rememberLastSession || !rememberLastUser)
                saveMemory();
        } catch (e) {
            log.warn("Failed to parse greetd memory:", e);
        }
    }

    function parseUserSessions(raw) {
        const sessions = {};
        if (!raw || typeof raw !== "object")
            return sessions;
        for (const username in raw) {
            const entry = raw[username];
            const sessionId = typeof entry?.sessionId === "string" ? entry.sessionId : "";
            if (!sessionId)
                continue;
            sessions[username] = {
                sessionId: sessionId,
                desktopId: typeof entry.desktopId === "string" ? entry.desktopId : ""
            };
        }
        return sessions;
    }

    function sessionForUser(username) {
        if (!rememberLastSession || !username)
            return null;
        return userSessions[username] || null;
    }

    function saveMemory() {
        let memory = {};
        if (rememberLastSession && lastSessionId)
            memory.lastSessionId = lastSessionId;
        if (rememberLastSession && lastSessionDesktopId)
            memory.lastSessionDesktopId = lastSessionDesktopId;
        if (rememberLastSession && Object.keys(userSessions).length > 0)
            memory.userSessions = userSessions;
        if (rememberLastUser && lastSuccessfulUser)
            memory.lastSuccessfulUser = lastSuccessfulUser;
        memoryFileView.setText(JSON.stringify(memory, null, 2));
    }

    function setLastSession(id, desktopId, username) {
        if (!rememberLastSession) {
            if (lastSessionId !== "" || lastSessionDesktopId !== "" || Object.keys(userSessions).length > 0) {
                lastSessionId = "";
                lastSessionDesktopId = "";
                userSessions = {};
                saveMemory();
            }
            return;
        }
        lastSessionId = id || "";
        lastSessionDesktopId = desktopId || "";
        if (!lastSessionId)
            lastSessionDesktopId = "";
        if (username && lastSessionId) {
            const sessions = Object.assign({}, userSessions);
            sessions[username] = {
                sessionId: lastSessionId,
                desktopId: lastSessionDesktopId
            };
            userSessions = sessions;
        }
        saveMemory();
    }

    function setLastSuccessfulUser(username) {
        if (!rememberLastUser) {
            if (lastSuccessfulUser !== "") {
                lastSuccessfulUser = "";
                saveMemory();
            }
            return;
        }
        lastSuccessfulUser = username || "";
        saveMemory();
    }

    FileView {
        id: memoryFileView
        path: root.memoryFile
        blockLoading: false
        blockWrites: false
        atomicWrites: true
        watchChanges: false
        printErrors: false
        onLoaded: {
            parseMemory(memoryFileView.text());
            root.memoryReady = true;
        }
        onLoadFailed: {
            root.memoryReady = true;
        }
    }
}
