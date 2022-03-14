pipeline {
  agent {
    node {
      label 'go'
    }
  }
  
    parameters {
        string(name:'APP_VERSION',defaultValue: '0.4.2',description:'')
        string(name:'CHART_VERSION',defaultValue: '0.4.2',description:'')
    }

    environment {
        // Docker access token,定义在凭证中 
        DOCKER_CREDENTIAL_ID = 'dockerhub-tkeel'
        // GitHub access token,定义在凭证中
        GITHUB_CREDENTIAL_ID = 'github'
        // k8s kubeconfig,定义在凭证中
        KUBECONFIG_CREDENTIAL_ID = 'kubeconfig'
        // Docker 仓库
        REGISTRY = 'docker.io'
        // Docker 空间
        DOCKERHUB_NAMESPACE = 'tkeelio'
        // Github 账号
        GITHUB_ACCOUNT = 'tkeel-io'
        // 组件名称
        APP_NAME = 'keel / rudder'
        // please ignore
        CHART_REPO_PATH = '/home/jenkins/agent/workspace/helm-charts'
    }

    stages {
        stage ('checkout scm') {
            steps {
                checkout(scm)
            }
        }

        stage ('build binary') {
            steps {
                container ('go') {
                    sh 'rm -rf /usr/local/go'
                    sh 'wget https://golang.org/dl/go1.17.6.linux-amd64.tar.gz'
                    sh 'tar -C /usr/local -xzf go1.17.6.linux-amd64.tar.gz'
                    sh 'go version'
                    sh 'make release GOOS=linux GOARCH=amd64'
                }
            }
        }        
 
        stage ('build & push image') {
            steps {
                container ('go') {
                    sh 'docker build -f docker/keel/Dockerfile -t $REGISTRY/$DOCKERHUB_NAMESPACE/keel:$BRANCH_NAME-$APP_VERSION ./dist/linux_amd64/release'
                    sh 'docker build -f docker/rudder/Dockerfile -t $REGISTRY/$DOCKERHUB_NAMESPACE/rudder:$BRANCH_NAME-$APP_VERSION ./dist/linux_amd64/release'
                    withCredentials([usernamePassword(passwordVariable : 'DOCKER_PASSWORD' ,usernameVariable : 'DOCKER_USERNAME' ,credentialsId : "$DOCKER_CREDENTIAL_ID" ,)]) {
                        sh 'echo "$DOCKER_PASSWORD" | docker login $REGISTRY -u "$DOCKER_USERNAME" --password-stdin'
                        sh 'docker push $REGISTRY/$DOCKERHUB_NAMESPACE/keel:$BRANCH_NAME-$APP_VERSION'
                        sh 'docker push $REGISTRY/$DOCKERHUB_NAMESPACE/rudder:$BRANCH_NAME-$APP_VERSION'
                    }
                }
            }
        }

        stage('build & push chart'){
          steps {
              container ('go') {
                sh 'helm3 package charts/keel --app-version=$APP_VERSION --version=$CHART_VERSION'
                sh 'helm3 package charts/rudder --app-version=$APP_VERSION --version=$CHART_VERSION'
                sh 'helm3 package charts/tkeel-middleware --app-version=$APP_VERSION --version=$CHART_VERSION'
                sh 'helm3 package charts/tkeel-plugin-components --app-version=$APP_VERSION --version=$CHART_VERSION'
                // input(id: 'release-image-with-tag', message: 'release image with tag?')
                  withCredentials([usernamePassword(credentialsId: "$GITHUB_CREDENTIAL_ID", passwordVariable: 'GIT_PASSWORD', usernameVariable: 'GIT_USERNAME')]) {
                    sh 'git config --global user.email "lunz1207@yunify.com"'
                    sh 'git config --global user.name "lunz1207"'
                    sh 'mkdir -p $CHART_REPO_PATH'
                    sh 'git clone https://$GIT_USERNAME:$GIT_PASSWORD@github.com/$GITHUB_ACCOUNT/helm-charts.git $CHART_REPO_PATH'
                    sh 'mv ./*.tgz $CHART_REPO_PATH/'
                    sh 'cd $CHART_REPO_PATH && helm3 repo index . --url=https://$GITHUB_ACCOUNT.github.io/helm-charts'
                    sh 'cd $CHART_REPO_PATH && git add . '
                    sh 'cd $CHART_REPO_PATH && git commit -m "feat:update chart"'
                    sh 'cd $CHART_REPO_PATH && git push https://$GIT_USERNAME:$GIT_PASSWORD@github.com/$GITHUB_ACCOUNT/helm-charts.git'
                  }
              }
          }
        }
    }
}