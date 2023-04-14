# Example 

```yaml

spec:
  selector:
    matchLabels:
      app: fortio
  

```

# Notes

- Restrict to recommended labels https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
  - app.kubernetes.io/name	- name of application 
  - app.kubernetes.io/instance	-  instance of application, if multiple independent install exist (multi-tenant)
  - app.kubernetes.io/version	
  - app.kubernetes.io/component	
  - app.kubernetes.io/part-of	 - canonical service ?
  - app.kubernetes.io/managed-by	
