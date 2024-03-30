#!/usr/bin/env bash
#

readarray -t fs < <(
    find . -name "*.sh" |
        xargs -I {} grep -E "^function (\w+)" {} | sort -u |
        sed -E "s/^function (\w+).*/\1/g"
)

readarray -t fsn < <(
    find . -name "*.sh" |
        xargs -I {} grep -E "^function (\w+)" {} | sort -u |
        sed -E "s/^function (\w+).*/\1/g" |
        sed -E "s/([A-Z])/_\L\1/g"
)

echo "#!/usr/bin/env bash" >script.sh
echo "source ~/.config/shell/functions.bash" >>script.sh
echo "gabyx::file_regex_replace -i '.*\.sh' . \\" >>script.sh

for i in "${!fs[@]}"; do
    f="${fs[$i]}"
    fn="${fsn[$i]}"
    cat <<EOF >>script.sh
    -r 's@$f@/$fn/g' \\
EOF

done
