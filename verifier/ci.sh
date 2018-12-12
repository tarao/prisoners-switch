on_exit() {
    [ "$imports_ok" = 1 ] || {
        json='{"success":false,"score":0,"steps":0,"used_switches":0,"message":"Violation: detected an illegal import line"}'
    }
    echo "verifier.result:$json"
    exit
}

[ "$CI" = 'true' ] && {
    trap on_exit EXIT
}
