FROM bitnami/minideb:bookworm

RUN apt-get update \
 && apt-get install -y --no-install-recommends wget procps ca-certificates \
 && rm -rf /var/lib/apt/lists/*


# Create directory structure to ensure mount points exist
# This is helpful locally and ensures permissions are clean
RUN mkdir -p /amux/docs \
    /amux/worktrees/jimmy \
    /amux/worktrees/frank \
    /amux/worktrees/greg \
    /amux/worktrees/curly \
    /amux/worktrees/jeremy \
    /amux/worktrees/big-pickle \
    /amux/worktrees/quaker \
    /amux/worktrees/quinton \
    /amux/journal

WORKDIR /amux

# Default command matches the script's sleep infinity behavior
CMD [ "sleep", "infinity" ]
