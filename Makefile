include .env

run:
	go build && ./sodan

list-model:
	curl https://api.openai.com/v1/models \
	-H "Authorization: Bearer $(OPENAI_API_KEY)" \
	| jq '.data | sort_by(.created) | map(.id)'