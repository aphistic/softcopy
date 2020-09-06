package frontend

//go:generate goblin -n frontend -r ../../../../../webui -i *.html -i *.elm -i **/*.elm
//go:generate goblin -n static -p frontend -r ../../../../../webui/static -i **/*
