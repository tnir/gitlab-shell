require_relative 'spec_helper'

require 'open3'

describe 'bin/gitlab-shell 2fa_recovery_codes' do
  include_context 'gitlab shell'

  def mock_server(server)
    server.mount_proc('/api/v4/internal/two_factor_recovery_codes') do |req, res|
      res.content_type = 'application/json'
      res.status = 200

      if req.query['key_id'] == '100'
        res.body = '{"success":true, "recovery_codes": ["1", "2"]}'
      else
        res.body = '{"success":false, "message": "Forbidden!"}'
      end
    end
  end

  shared_examples 'dialog for regenerating recovery keys' do
    context 'when runs successfully' do
      let(:cmd) { "#{gitlab_shell_path} key-100" }

      context 'when the user agrees to regenerate keys' do
        it 'the recovery keys are regenerated' do
          Open3.popen2(env, cmd) do |stdin, stdout|
            expect(stdout.gets).to eq("Are you sure you want to generate new two-factor recovery codes?\n")
            expect(stdout.gets).to eq("Any existing recovery codes you saved will be invalidated. (yes/no)\n")

            stdin.puts('yes')

            expect(stdout.flush.read).to eq(
              "\nYour two-factor authentication recovery codes are:\n\n" \
              "1\n2\n\n" \
              "During sign in, use one of the codes above when prompted for\n" \
              "your two-factor code. Then, visit your Profile Settings and add\n" \
              "a new device so you do not lose access to your account again.\n"
            )
          end
        end
      end

      context 'when the user disagrees to regenerate keys' do
        it 'the recovery keys are not regenerated' do
          Open3.popen2(env, cmd) do |stdin, stdout|
            expect(stdout.gets).to eq("Are you sure you want to generate new two-factor recovery codes?\n")
            expect(stdout.gets).to eq("Any existing recovery codes you saved will be invalidated. (yes/no)\n")

            stdin.puts('no')

            expect(stdout.flush.read).to eq(
              "\nNew recovery codes have *not* been generated. Existing codes will remain valid.\n"
            )
          end
        end
      end
    end

    context 'when API error occurs' do
      let(:cmd) { "#{gitlab_shell_path} key-101" }

      context 'when the user agrees to regenerate keys' do
        it 'the recovery keys are regenerated' do
          Open3.popen2(env, cmd) do |stdin, stdout|
            expect(stdout.gets).to eq("Are you sure you want to generate new two-factor recovery codes?\n")
            expect(stdout.gets).to eq("Any existing recovery codes you saved will be invalidated. (yes/no)\n")

            stdin.puts('yes')

            expect(stdout.flush.read).to eq("\nAn error occurred while trying to generate new recovery codes.\nForbidden!\n")
          end
        end
      end
    end
  end

  let(:env) { {'SSH_CONNECTION' => 'fake', 'SSH_ORIGINAL_COMMAND' => '2fa_recovery_codes' } }

  describe 'without go features' do
    before(:context) do
      write_config(
        "gitlab_url" => "http+unix://#{CGI.escape(tmp_socket_path)}",
      )
    end

    it_behaves_like 'dialog for regenerating recovery keys'
  end

  describe 'with go features' do
    before(:context) do
      write_config(
        "gitlab_url" => "http+unix://#{CGI.escape(tmp_socket_path)}",
        "migration" => { "enabled" => true,
                        "features" => ["2fa_recovery_codes"] }
      )
    end

    it_behaves_like 'dialog for regenerating recovery keys'
  end
end
