PROJECT=wlhe-cr

run:
	gcloud  run deploy istiod --allow-unauthenticated \
	--image gcr.io/${PROJECT}/pilot:gcp \
     --args=--log_output_level=debug \
	--port 8080 --memory 2G --platform managed --project ${PROJECT} --region us-central1

#	gcloud run deploy istiod-xds --allow-unauthenticated \
#		--image gcr.io/${PROJECT}/pilot:gcp \
#        --args=--log_output_level=debug \
#    	--port 15010 --memory 2G -platform managed --project ${PROJECT}

run1:
	PROJECT=wlhe-cr make run
