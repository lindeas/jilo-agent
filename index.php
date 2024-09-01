<?php

require 'config.php';
include 'functions.php';

$scriptname = basename($_SERVER['SCRIPT_NAME']);
$request = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);

// the response is in JSON
header('Content-Type: application/json');

// nginx status
if ($request === '/nginx' || $request === "/$scriptname/nginx") {
    $data = [
        'nginx_status'		=> getNginxStatus(),
        'nginx_connections'	=> getNginxConnections(),
    ];
    echo json_encode($data, JSON_PRETTY_PRINT) . "\n";

// prosody status
} elseif ($request === '/prosody' || $request === "/$scriptname/prosody") {
    $data = [
        'prosody_status'	=> getProsodyStatus(),
    ];
    echo json_encode($data, JSON_PRETTY_PRINT) . "\n";

// jicofo status
} elseif ($request === '/jicofo' || $request === "/$scriptname/jicofo") {
    $jicofoStatsCommand = "curl -s $jicofoStatsURL";
    $jicofoStatsData = getJicofoStats($jicofoStatsCommand);
    $data = [
        'jicofo_status'		=> getJicofoStatus(),
        'jicofo_API_stats'	=> $jicofoStatsData,
    ];
    echo json_encode($data, JSON_PRETTY_PRINT) . "\n";

// jvb status
} elseif ($request === '/jvb' || $request === "/$scriptname/jvb") {
    $jvbStatsCommand = "curl -s $jvbStatsURL";
    $jvbStatsData = getJVBStats($jvbStatsCommand);
    $data = [
        'jvb_status'		=> getJVBStatus(),
        'jvb_API_stats'		=> $jvbStatsData,
    ];
    echo json_encode($data, JSON_PRETTY_PRINT) . "\n";

// default response - error
} else {
    echo json_encode(['error' => 'Endpoint not found']) . "\n";
}

?>
