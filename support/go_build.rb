# Helper functions to build go code in gitlab-shell

require 'fileutils'

# This will set the ROOT_PATH variable
require_relative '../lib/gitlab_init'

module GoBuild
  GO_DIR = 'go'.freeze
  BUILD_DIR = File.join(ROOT_PATH, 'go_build')

  GO_ENV = {
    'GOBIN' => File.join(BUILD_DIR, 'bin'),
    'GO111MODULE' => 'on',
    'GOPROXY' => 'https://proxy.golang.org'
  }.freeze

  def ensure_build_dir_exists
    FileUtils.mkdir_p(BUILD_DIR)
  end

  def run!(env, cmd, options = {})
    raise "env must be a hash" unless env.is_a?(Hash)
    raise "cmd must be an array" unless cmd.is_a?(Array)

    unless system(env, *cmd, options)
      abort "command failed: #{env.inspect} #{cmd.join(' ')}"
    end
  end
end
