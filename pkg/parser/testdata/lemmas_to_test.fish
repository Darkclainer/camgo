#!/usr/bin/env fish

if test -z "$argv[1]" 
    echo You should specify input html file
    exit 1
end

camgo -f "$argv[1]" | jq 'keys[] as $i | {index: $i, lemma: .[$i]}' | jq --slurp '{length: . | length, type: "lemmas", content: .}'
