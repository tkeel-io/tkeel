pipeline {
  agent {
    node {
      label 'base'
    }
  }
    parameters {
        string(name:'PROD',defaultValue: 'no',description:'是否')

        string(name:'REGISTRY',defaultValue: 'harbor.wuxs.vip:30003',description:'组织')
        string(name:'REPOSITORY',defaultValue: 'tkeel-io',description:'仓库')

        string(name:'DOCKER_IMAGE_TAG',defaultValue: '0.1.0',description:'镜像版本')
        string(name:'HELM_CHART_VERSION',defaultValue: '0.1.0',description:'插件版本')

    }

    environment {
        /*
        应用变量
        */
        APP_NAME = 'tkeel-middleware'
        
        /*
        凭证信息
        */
        PRIVATE_REPO_CREDENTIAL_ID = 'harbor'
        KUBECONFIG_CREDENTIAL_ID = 'kubeconfig'

        /*
        配置,从凭证中读取
        */
        PRIVATE_REPO_CONFIG = 'private-repo'
        PLUGIN_CONFIG = 'tkeel-plugin-config'
    }

    stages {
        stage ('checkout scm') {
            steps {
                checkout(scm)
            }
        }

        stage('get config'){
            /*
            从凭证中读取配置
            */
            steps {
                container ('base'){
                    withCredentials([usernamePassword(credentialsId: "$PRIVATE_REPO_CONFIG", usernameVariable: 'registry',passwordVariable: 'repository')]) {
                        script {
                            env.REGISTRY = registry
                            env.REPOSITORY = repository
                            }
                        }

                    withCredentials([usernamePassword(credentialsId: "$PRIVATE_REPO_CREDENTIAL_ID", usernameVariable: 'username',passwordVariable: 'password')]) {
                        script {
                            env.USERNAME = username
                            env.PASSWORD = password
                            }
                        }

                    withCredentials([usernamePassword(credentialsId: "$PLUGIN_CONFIG", usernameVariable: 'enable_config',passwordVariable: 'config')]) {
                        script {
                            env.TKEEL_PLUGIN_ENABLE_UPGRADE = enable_config
                            env.TKEEL_PLUGIN_CONFIG = config
                            }
                        }
                    }
                }
            }

        stage('set env'){
            environment {
                COMMIT_ID = "${sh(script:'git rev-parse --short HEAD',returnStdout:true)}"
                TIMESTAMP = "${sh(script:'date -d "+8 hour" "+%m.%d.%H%M%S"',returnStdout:true)}"
            }
            steps {
                container ('base'){
                    script {
                        if (params.PROD == "yes"){
                            /*
                            重写 REGISTRY & REPOSITORY
                            */
                            env.REGISTRY = params.REGISTRY
                            env.REPOSITORY = params.REPOSITORY
                            env.DOCKER_IMAGE_TAG = params.DOCKER_IMAGE_TAG
                            env.DOCKER = env.REGISTRY + "/" + env.REPOSITORY + "/" + env.APP_NAME  + ":" + env.DOCKER_IMAGE_TAG
                            env.CHART = params.HELM_CHART_VERSION
                            env.TKEEL_PLUGIN_ENABLE_UPGRADE = 'yes'
                        }else{
                            env.DOCKER_IMAGE_TAG = env.COMMIT_ID
                            env.DOCKER =  env.REGISTRY + "/" + env.REPOSITORY + "/" + env.APP_NAME  + ":" + env.DOCKER_IMAGE_TAG
                            env.CHART = env.TIMESTAMP
                        }
                    }
                }
            }
        }

        stage('build & push') {
            environment {
                /*
                helm 环境变量
                */
                HELM_EXPERIMENTAL_OCI=1

                /*
                chart 文件夹相对路径
                */
                CHART_PATH = 'charts/tkeel-middleware'
            }
            steps {
                container ('base') {
                    /*
                    helm chart
                    */
                    sh 'helm3 plugin install https://github.com/chartmuseum/helm-push'
                    sh 'helm3 registry login -u $USERNAME -p $PASSWORD $REGISTRY'
                    sh 'helm3 package $CHART_PATH --version=$CHART --app-version=$DOCKER_IMAGE_TAG' 
                    sh 'helm3 cm-push $APP_NAME-*.tgz https://$REGISTRY/chartrepo/$REPOSITORY --username=$USERNAME --password=$PASSWORD'
                }
            }
        }
 
        stage('install or upgrade plugin') {
            steps {
                container ('base') {
                    withCredentials([kubeconfigFile(credentialsId: env.KUBECONFIG_CREDENTIAL_ID,variable: 'KUBECONFIG')]) {
                        script {                        
                            if (env.TKEEL_PLUGIN_ENABLE_UPGRADE == 'yes'){
                                sh 'wget -q https://raw.githubusercontent.com/tkeel-io/cli/master/install/install.sh -O - | /bin/bash'
                                sh 'tkeel admin login -p changeme'
                                sh 'mv devops/config/$TKEEL_PLUGIN_CONFIG ~/.tkeel/config.yaml'
                                sh 'tkeel upgrade --repo-url=https://$REGISTRY/chartrepo/$REPOSITORY --repo-name=$REPOSITORY  ----middleware-version=$CHART'
                            }else{
                                sh 'echo do not install or upgrade'
                            }
                        }
                    }                
                }
            }
        }
    }
}