Feature:
  In order to be able to access the application with a proper name
  As a developer
  I want to be able to chose the domain names that will be handled by the vamp router

  Scenario: kubernetesReverseProxy compatibility
    Given the k8s service "app" is in the namespace "default"
    And the k8s service "app" has the following annotations:
      | name                   | value                                              |
      | kubernetesReverseproxy | {"hosts": [{"host": "example.com", "port": "80"}]} |
    And the k8s service "app" IP is "1.2.3.4"
    When the k8s service named "app" is created
    And the vamp service "example.com" should be created
    And the vamp service "example.com" should only contain the backend "1.2.3.4"
