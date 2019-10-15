pipeline {
    agent any

    environment {
        imageName = "rediseen"
        tag = "latest"
    }

    stages {
        stage('Check CICD Server Status') {
            steps {
                sh '''uptime'''

                sh '''df -h'''
                sh '''docker image ls --all'''

                sh '''docker image prune -f'''

                sh '''df -h'''
                sh '''docker image ls --all'''
            }
        }
        stage('Build') {
            steps {
                script {
                    sh '''docker build -t ${imageName}:${tag} .'''
                }
            }
        }
     }
}