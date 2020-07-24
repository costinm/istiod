
IMAGE=pilot:gcp
PROJECT_ID=costin-asm1

GAC=$HOME/.config/gcloud/legacy_credentials/costin@google.com/adc.json

#--entrypoint /bin/sh \

PORT=8080 && docker run -it --rm --name istiod \
-p 9090:${PORT} \
-e PORT=${PORT} \
-e K_SERVICE=dev \
-e K_CONFIGURATION=dev \
-e K_REVISION=dev-00001 \
-e CLOUDSDK_AUTH_CREDENTIAL_FILE_OVERRIDE=/var/run/secrets/google/google.json \
-e GOOGLE_APPLICATION_CREDENTIALS=/var/run/secrets/google/google.json \
-v $GAC:/var/run/secrets/google/google.json:ro \
gcr.io/$PROJECT_ID/$IMAGE $*
