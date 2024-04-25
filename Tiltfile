# -*- mode: Python -*-

docker_repo = 'testinbox'

docker_build(docker_repo, '.', dockerfile='Dockerfile')

k8s_yaml('.deploy/local.yaml')
k8s_resource('testinbox', port_forwards='8080:8080')
