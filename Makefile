test:
	for i in *; do test -d $$i || continue; cd $$i || continue; \
		echo Testing $$i; \
		go test -v -count=1 -cover ./...; \
		cd ..; \
		echo Done $$i; \
	done
