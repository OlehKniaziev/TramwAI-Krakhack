FROM ollama/ollama:latest
WORKDIR /llm
COPY ./init.sh .
RUN chmod +x /llm/init.sh
RUN /llm/init.sh
