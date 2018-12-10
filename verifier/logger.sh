msg() {
    _msg_command=$(basename $0 .sh)
    _msg_level=$(echo "$1" | tr '[a-z]' '[A-Z]'); shift
    _msg_action="$1"; shift;
    [ -n "$_msg_action" ] && _msg_action=".$_msg_action"
    _msg_time=$(date --iso-8601=seconds)
    _msg_tag=$(echo "$_msg_level" | cut -c 1-1)

    _msg="$_msg_tag [$_msg_time #$$] $_msg_command$_msg_action : $@"

    test -t 2 || {
        echo "$_msg" >&2
        return
    }

    case "$_msg_level" in
    FATAL|ERROR)
        echo "[31m$_msg[0m" >&2
        ;;
    WARN)
        echo "[33m$_msg[0m" >&2
        ;;
    INFO)
        echo "$_msg" >&2
        ;;
    esac
}

fatal() {
    [ 0 -lt "$verbose_level" ] && msg fatal '' "$@"
}

error() {
    [ 1 -lt "$verbose_level" ] && msg error '' "$@"
}

warn() {
    [ 2 -lt "$verbose_level" ] && msg warn '' "$@"
}

info() {
    [ 3 -lt "$verbose_level" ] && msg info "$@"
}

debug() {
    [ 4 -lt "$verbose_level" ] && msg debug "$@"
}

die() {
    fatal "$@"
    exit 1
}

VERBOSE_LEVEL_FATAL=1
VERBOSE_LEVEL_ERROR=2
VERBOSE_LEVEL_WARN=3
VERBOSE_LEVEL_INFO=4
VERBOSE_LEVEL_DEBUG=5
