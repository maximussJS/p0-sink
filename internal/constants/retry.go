package constants

import "time"

const PIPELINE_RETRY_DELAY = 5 * time.Second
const PIPELINE_RETRY_MAX_ATTEMPTS = 5
const PIPELINE_RETRY_ATTEMPTS_RESET_INTERVAL = 180000
