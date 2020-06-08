FROM istionightly/base_debug

ADD fortio /usr/bin/fortio
RUN mkdir /usr/share/fortio
ADD ui/static /usr/share/fortio/static
ADD ui/templates /usr/share/fortio/templates

ENTRYPOINT ["/usr/bin/fortio"]
