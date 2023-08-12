.PHONY: deploy

deploy:
	@echo "Deploying helloworld service on Cloud Run..."
	@gcloud run deploy helloworld --memory=128Mi --max-instances=1 --source .
