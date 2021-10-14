# bamboo-cctray

Inspired by https://github.com/chadlwilson/bamboo_cctray_proxy. But written in Golang to enable easy running on cross platform.

## Usage
1. Download latest builds for your platform from Actions tab (e.g. https://github.com/ruqqq/bamboo-cctray/actions/runs/894236499)
2. Create `bamboo.yml`:
```yaml
  - bamboo1:
      url: https://some-bamboo.com
      basic_auth:
        username: user
        password: password
      build_keys:
        - PROJECT1-PLAN1
        - PROJECT2-PLAN2

  - bamboo2:
      url: https://some-bamboo.com
      basic_auth:
        username: user
        password: password
      projects:
        - PROJECT1
```
3. Run `bamboo-cctray bamboo.yml` from the extracted archive in step 1
4. Open browser to `http://localhost:7000/dashboard/cctray.xml`

## TODO
- Add tests
- Support running server on different port
- Remove dependency on go-bamboo
- Publish release (artifacts expires!)
