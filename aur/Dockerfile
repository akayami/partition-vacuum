FROM archlinux:base-devel

# Install dependencies
RUN pacman -Syu --noconfirm go git pacman-contrib

# Create a non-root user for makepkg
RUN useradd -m builder && \
    echo "builder ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/builder

# Set up work directory
WORKDIR /home/builder/package
COPY . .
RUN chown -R builder:builder /home/builder/package

# Switch to builder user
USER builder

# Entrypoint script to handle versioning and building
COPY aur/entrypoint.sh /entrypoint.sh
USER root
RUN chmod +x /entrypoint.sh
USER builder

ENTRYPOINT ["/entrypoint.sh"]
