pragma Singleton
pragma ComponentBehavior: Bound
import QtQuick
import Quickshell

Singleton {
    id: root

    property string passwordBuffer: ""
    property string username: ""
    property string usernameInput: ""
    property bool showPasswordInput: false
    property string selectedSession: ""
    property string selectedSessionPath: ""
    property string selectedSessionDesktopId: ""
    property string selectedSessionDesktopNames: ""
    property string pamState: ""
    property bool unlocking: false
    property bool awaitingExternalAuth: false
    property bool pendingPasswordResponse: false
    property bool passwordSubmitRequested: false
    property string authPromptText: ""
    property string authFeedbackMessage: ""

    property var sessionList: []
    property var sessionExecs: []
    property var sessionPaths: []
    property var sessionDesktopIds: []
    property var sessionDesktopNames: []
    property int currentSessionIndex: 0
    property bool sessionManuallySelected: false

    function reset() {
        showPasswordInput = false;
        username = "";
        usernameInput = "";
        passwordBuffer = "";
        pamState = "";
        awaitingExternalAuth = false;
        pendingPasswordResponse = false;
        passwordSubmitRequested = false;
        authPromptText = "";
        authFeedbackMessage = "";
    }
}
