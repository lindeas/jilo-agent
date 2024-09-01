<?php

// get nginx data
function getNginxStatus() {
    $status = trim(shell_exec('systemctl is-active nginx'));
    return ($status === 'active') ? 'running' : 'not running';
}
function getNginxConnections() {
    $connections = shell_exec("netstat -an | grep ':80' | wc -l");
    return intval(trim($connections));
}


// get prosody data
function getProsodyStatus() {
    $status = trim(shell_exec('systemctl is-active prosody'));
    return ($status === 'active') ? 'running' : 'not running';
}


// get jicofo data
function getJicofoStatus() {
    $status = trim(shell_exec('systemctl is-active jicofo'));
    return ($status === 'active') ? 'running' : 'not running';
}


// get JVB data
function getJVBStatus() {
    $status = trim(shell_exec('systemctl is-active jitsi-videobridge'));
    return ($status === 'active') ? 'running' : 'not running';
}

?>
