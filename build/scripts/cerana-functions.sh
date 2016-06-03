#+
# Some standard functions for Mistify-OS scripts.
#-
cmdline="$0 $*"
projectdir=$PWD    # Save this directory for later.

testceranastatedir=$projectdir/.testcerana
if [ ! -e $testceranastatedir ]; then
    mkdir -p $testceranastatedir
fi

# Which branch this script is running with.
if [ -e .git ]; then
    ceranatestbranchdefault=`git symbolic-ref -q --short HEAD`
    # Jenkins detaches for branches so need to use a commit ID instead.
    if [ -z "$ceranatestbranchdefault" ]; then
        ceranatestbranchdefault=`git rev-parse HEAD`
    fi
fi

function get_test_default() {
    # Parameters:
    #   1: option name
    #   2: default value
    if [ -e $testceranastatedir/$1 ]; then
        r=`cat $testceranastatedir/$1`
    else
        r=$2
    fi
    verbose The test default for $1 is $2
    echo $r
}

function set_test_default() {
    # Parameters:
    #   1: option name
    #   2: value
    echo "$2">$testceranastatedir/$1
    verbose The test default $1 has been set to $2
}

function reset_test_default() {
    # Parameters:
    #   1: option name
    if [ -e $testceranastatedir/$1 ]; then
        rm $testceranastatedir/$1
        verbose Option $1 test default has been reset.
    else
        verbose Option $1 test default has not been set.
    fi
}

function get_build_default() {
    # Parameters:
    #   1: state directory
    #   2: option name
    #   3: default value
    if [ -e $1/$2 ]; then
        r=`cat $1/$2`
    else
        r=$3
    fi
    verbose The default for $2 is $3
    echo $r
}

function clear_test_variable() {
    # Parameters:
    #   1: variable name and default value pair delimited by the delimeter (2)
    #   2: an optional delimeter character (defaults to '=')
    if [ -z "$2" ]; then
        d='='
    else
        d=$2
    fi
    e=(`echo "$1" | tr "$d" " "`)
    verbose ""
    verbose Clearing state variable: ${e[0]}
    reset_test_default ${e[0]}
}

function init_test_variable() {
    # Parameters:
    #   1: variable name and default value pair delimited by the delimeter (2)
    #   2: an optional delimeter character (defaults to '=')
    if [ ! -z "$resetdefaults" ]; then
        clear_test_variable $v
    fi
    if [ -z "$2" ]; then
        d='='
    else
        d=$2
    fi
    e=(`echo "$1" | tr "$d" " "`)
    verbose ""
    verbose State variable default: "${e[0]} = ${e[1]}"
    eval val=\$${e[0]}
    if [ -z "$val" ]; then
        verbose Setting ${e[0]} to default: ${e[1]}
        eval ${e[0]}=$(get_test_default ${e[0]} ${e[1]})
    else
        if [ "$val" = "default" ]; then
            verbose Setting ${e[0]} to default: ${e[1]}
            eval ${e[0]}=${e[1]}
        else
            eval ${e[0]}=$val
        fi
    fi
    eval val=\$${e[0]}
    verbose "State variable: ${e[0]} = $val"
    verbose Saving current settings.
    set_test_default ${e[0]} $val
}

function get_ceranaos_version() {
    # Parameters:
    #    1: ceranaosdir - Where the cerana-os repo clone is located.
    if [ -d $1 ]; then
        pushd $1 >/dev/null
        # Which branch of the cerana-os is being checked.
        r=`git symbolic-ref -q --short HEAD`
        # Jenkins detaches for branches so need to use a commit ID instead.
        if [ -z "$r" ]; then
            r=`git rev-parse HEAD`
        fi
        popd >/dev/null
    else
        r=master
    fi
    echo $r
}

greybg='\e[47m'
darkgreybg='\e[100m'
green='\e[0;32m'$darkgreybg
yellow='\e[0;33m'$greybg
red='\e[0;31m'$greybg
blue='\e[0;34m'$darkgreybg
lightblue='\e[1;34m'$greybg
white='\e[1;37m'$darkgreybg
nc='\e[0m'
id=$(basename $0)

message () {
    echo -e "$green$id$nc: $*"
}

tip () {
    echo -e "$green$id$nc: $white$*$nc"
}

warning () {
    echo -e "$green$id$yellow WARNING$nc: $*"
}

error () {
    echo >&2 -e "$green$id$red ERROR$nc: $*"
}

verbose () {
    if [[ "$verbose" == "y" ]]; then
        echo >&2 -e "$lightblue$id$nc: $*"
    fi
}

log () {
    echo $* >>$testlogdir/$testlog
    verbose "$*"
}

function die() {
    error "$@"
    exit 1
}

function run() {
    verbose "Running: '$@'"
    "$@"; code=$?; [ $code -ne 0 ] && die "Command [$*] failed with status code $code";
    return $code
}

function run_ignore {
    verbose "Running: '$@'"
    "$@"; code=$?; [ $code -ne 0 ] && verbose "Command [$*] returned status code $code";
    return $code
}

function confirm () {
    read -r -p "${1:-Are you sure? [y/N]} " response
    case $response in
    [yY][eE][sS]|[yY])
        true
        ;;
    *)
        false
        ;;
    esac
}

is_mounted () {
    mount | grep $1
    return $?
}
