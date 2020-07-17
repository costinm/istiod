Based on the [Bookinfo guide](https://istio.io/docs/guides/bookinfo.html) - with changes for 
kustomize and the a-la-carte intsaller.

Changes compared with the guide:

- bookinfo is installed using kustomize (`kubectl apply -k`)
- the installation uses 'bookinfo' namespace

TODO: update the guide for the new installer/templates.
