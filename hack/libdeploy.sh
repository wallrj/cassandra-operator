# Some useful bash functions from https://github.com/jetstack/navigator/blob/master/hack/libe2e.sh

function retry() {
    local TIMEOUT=60
    local SLEEP=10
    while :
    do
        case "${1}" in
            TIMEOUT=*|SLEEP=*)
                local "${1}"
                shift
                ;;
            *)
                break
                ;;
        esac
    done
    local start_time
    start_time="$(date +"%s")"
    local end_time
    end_time="$(($start_time + ${TIMEOUT}))"
    until "${@}"
    do
        local exit_code="${?}"
        local current_time="$(date +"%s")"
        local remaining_time="$((end_time - current_time))"
        if [[ "${remaining_time}" -le 0 ]]; then
            return "${exit_code}"
        fi
        local sleep_time="${SLEEP}"
        if [[ "${remaining_time}" -lt "${SLEEP}" ]]; then
            sleep_time="${remaining_time}"
        fi
        sleep "${sleep_time}"
    done
}
