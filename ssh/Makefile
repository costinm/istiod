
ssh/manual:
	ssh-keygen -t ecdsa  -f id_ecdsa -N ""


	ssh -v  -p 15022 -F /dev/null \
		-o "UserKnownHostsFile /dev/null" \
		-o StrictHostKeyChecking=no \
		-o CertificateFile=${HOME}/.ssh/id_ecdsa-cert-big1.pub localhost
