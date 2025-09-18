# For example
#	goawk -f latest-notes.awk release_note_*.md | gh release create -d --notes-file - -t $(VERSION) $(VERSION) $(wildcard $(NAME)-$(VERSION)-*.zip)

match($0,/^v[0-9]+\.[0-9]+\.[0-9]+$/) > 0 {
    flag = ++f[FILENAME]
    if ( flag == 1 ) {
        version = substr($0,RSTART,RLENGTH)
        printf "\n### Changes in %s",version
        if (FILENAME ~ /ja/) {
            print " (Japanese)"
        } else if (FILENAME ~ /en/ ){
            print " (English)"
        } else {
            print ""
        }
    }
}

f[FILENAME]==1 && /^$/{
    section++
}

(f[FILENAME]==1 && section %2 == 1 ){
    print
}
