pipeline {
    agent any
    
    environment {
        DOCKER_REGISTRY = 'your-registry.amazonaws.com'
        KUBECONFIG = '/var/jenkins_home/.kube/config'
        SLACK_WEBHOOK = credentials('slack-webhook')
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
                script {
                    env.GIT_COMMIT_SHORT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    env.BUILD_TIMESTAMP = sh(script: 'date +%Y%m%d_%H%M%S', returnStdout: true).trim()
                }
            }
        }
        
        stage('Build') {
            parallel {
                stage('API Gateway') {
                    steps {
                        dir('gateway/api-gateway') {
                            sh 'docker build -t ${DOCKER_REGISTRY}/api-gateway:${GIT_COMMIT_SHORT} .'
                        }
                    }
                }
                
                stage('Services') {
                    steps {
                        script {
                            def services = ['auth-service', 'vehicle-service', 'driver-service', 'trip-service', 
                                          'maintenance-service', 'tracking-service', 'alert-service', 'geofencing-service']
                            
                            services.each { service ->
                                dir("services/${service}") {
                                    sh "docker build -t \${DOCKER_REGISTRY}/${service}:${GIT_COMMIT_SHORT} ."
                                }
                            }
                        }
                    }
                }
            }
        }
        
        stage('Test') {
            parallel {
                stage('Unit Tests') {
                    steps {
                        script {
                            def services = ['auth-service', 'vehicle-service', 'driver-service', 'trip-service', 
                                          'maintenance-service', 'tracking-service', 'alert-service', 'geofencing-service']
                            
                            services.each { service ->
                                dir("services/${service}") {
                                    sh 'go test -v ./...'
                                }
                            }
                        }
                    }
                }
                
                stage('Integration Tests') {
                    steps {
                        sh './scripts/run-integration-tests.sh'
                    }
                }
                
                stage('Security Scan') {
                    steps {
                        sh 'trivy fs . --exit-code 1 --severity HIGH,CRITICAL'
                    }
                }
            }
        }
        
        stage('Push Images') {
            steps {
                script {
                    withCredentials([usernamePassword(credentialsId: 'docker-registry', 
                                                     usernameVariable: 'DOCKER_USER', 
                                                     passwordVariable: 'DOCKER_PASS')]) {
                        sh 'docker login -u $DOCKER_USER -p $DOCKER_PASS $DOCKER_REGISTRY'
                        
                        def services = ['api-gateway', 'auth-service', 'vehicle-service', 'driver-service', 
                                      'trip-service', 'maintenance-service', 'tracking-service', 
                                      'alert-service', 'geofencing-service']
                        
                        services.each { service ->
                            sh "docker push \${DOCKER_REGISTRY}/${service}:${GIT_COMMIT_SHORT}"
                            sh "docker tag \${DOCKER_REGISTRY}/${service}:${GIT_COMMIT_SHORT} \${DOCKER_REGISTRY}/${service}:latest"
                            sh "docker push \${DOCKER_REGISTRY}/${service}:latest"
                        }
                    }
                }
            }
        }
        
        stage('Deploy to Staging') {
            when {
                branch 'develop'
            }
            steps {
                script {
                    def services = ['api-gateway', 'auth-service', 'vehicle-service', 'driver-service', 
                                  'trip-service', 'maintenance-service', 'tracking-service', 
                                  'alert-service', 'geofencing-service']
                    
                    dir('infrastructure/helm/flyfleet-service') {
                        sh 'kubectl config use-context staging'
                        services.each { service ->
                            sh """
                                helm upgrade --install ${service} . \\
                                  --namespace flyfleet-staging \\
                                  --create-namespace \\
                                  --set image.repository=\${DOCKER_REGISTRY}/${service} \\
                                  --set image.tag=\${GIT_COMMIT_SHORT}
                            """
                        }
                    }
                }
            }
        }
        
        stage('Run Staging Tests') {
            when {
                branch 'develop'
            }
            steps {
                sh './scripts/run-staging-tests.sh'
            }
        }
        
        stage('Deploy to Production') {
            when {
                branch 'main'
            }
            steps {
                timeout(time: 30, unit: 'MINUTES') {
                    input message: 'Deploy to Production?', ok: 'Deploy'
                }
                
                script {
                    def services = ['api-gateway', 'auth-service', 'vehicle-service', 'driver-service', 
                                  'trip-service', 'maintenance-service', 'tracking-service', 
                                  'alert-service', 'geofencing-service']
                    
                    dir('infrastructure/helm/flyfleet-service') {
                        sh 'kubectl config use-context production'
                        services.each { service ->
                            sh """
                                helm upgrade --install ${service} . \\
                                  --namespace flyfleet-prod \\
                                  --create-namespace \\
                                  --set image.repository=\${DOCKER_REGISTRY}/${service} \\
                                  --set image.tag=\${GIT_COMMIT_SHORT}
                            """
                        }
                    }
                }
            }
        }
        
        stage('Post Deploy Verification') {
            steps {
                sh './scripts/post-deploy-verification.sh'
            }
        }
    }
    
    post {
        success {
            script {
                sh "curl -X POST -H 'Content-type: application/json' --data '{\"text\":\"✅ Build Success: ${env.JOB_NAME} #${env.BUILD_NUMBER}\"}' ${SLACK_WEBHOOK}"
            }
        }
        
        failure {
            script {
                sh "curl -X POST -H 'Content-type: application/json' --data '{\"text\":\"❌ Build Failed: ${env.JOB_NAME} #${env.BUILD_NUMBER}\"}' ${SLACK_WEBHOOK}"
            }
        }
        
        cleanup {
            sh 'docker system prune -f'
        }
    }
}