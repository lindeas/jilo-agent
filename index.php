<?php

include 'functions.php';

// the response is in JSON
header('Content-Type: application/json');

// nginx status
if ($_SERVER['REQUEST_URI'] === '/nginx') {
    $data = [
        'nginx_status'		=> getNginxStatus(),
        'nginx_connections'	=> getNginxConnections(),
    ];
    echo json_encode($data, JSON_PRETTY_PRINT) . "\n";

// jvb status
} elseif ($_SERVER['REQUEST_URI'] === '/jvb') {
    $data = [
        'jvb_status'		=> getJVBStatus(),
    ];
    echo json_encode($data, JSON_PRETTY_PRINT) . "\n";

// default response - error
} else {
    echo json_encode(['error' => 'Endpoint not found']) . "\n";
}

?>
