<?php

// nginx status
if ($_SERVER['REQUEST_URI'] === '/nginx/status') {
    echo json_encode(['nginx' => shell_exec('/etc/init.d/nginx status')]);

// jvb status
} elseif ($_SERVER['REQUEST_URI'] === '/jvb/status') {
    echo json_encode(['jvb' => shell_exec('/etc/init.d/jitsi-videobridge status')]);

// default response - error
} else {
    echo json_encode(['error' => 'Endpoint not found']);
}

?>
