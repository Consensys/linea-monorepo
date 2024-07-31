if [ ! -f .env ]
then
    echo ".env is not initialized, creating one"
    SUPER_SECRET="$(date '+%s')"
    echo "WS_SECRET=\"$SUPER_SECRET\"" > .env
    echo "SECRET_KEY_BASE=\"$SUPER_SECRET\"" >> .env
fi
