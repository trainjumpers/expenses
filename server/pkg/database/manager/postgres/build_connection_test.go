package postgres

import (
	"net/url"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TestBuildConnectionString", func() {
	var (
		host     string
		user     string
		password string
		dbname   string
		port     int
		schema   string
		sslmode  string
		config   *PoolConfig

		connStr string
		parsed  *url.URL
	)

	BeforeEach(func() {
		host = "localhost"
		user = "test_user"
		password = "p@ss/w:rd?&="
		dbname = "testdb"
		port = 5432
		schema = "public"
		sslmode = "require"

		config = &PoolConfig{
			ApplicationName:      "my-app",
			MaxConns:             20,
			MinConns:             5,
			MaxConnLifetime:      30 * time.Minute,
			MaxConnIdleTime:      5 * time.Minute,
			HealthCheckPeriod:    1 * time.Minute,
			PreferSimpleProtocol: false,
		}

		connStr = buildConnectionString(
			host,
			user,
			password,
			dbname,
			port,
			schema,
			sslmode,
			config,
		)

		var err error
		parsed, err = url.Parse(connStr)
		Expect(err).NotTo(HaveOccurred())
	})

	It("uses the postgresql scheme", func() {
		Expect(parsed.Scheme).To(Equal("postgresql"))
	})

	It("sets host and port correctly", func() {
		Expect(parsed.Host).To(Equal("localhost:5432"))
	})

	It("sets the database name as the path", func() {
		Expect(parsed.Path).To(Equal("/testdb"))
	})

	It("includes the username", func() {
		Expect(parsed.User.Username()).To(Equal("test_user"))
	})

	It("percent-encodes the password correctly", func() {
		pw, ok := parsed.User.Password()
		Expect(ok).To(BeTrue())

		decoded, err := url.PathUnescape(pw)
		Expect(err).NotTo(HaveOccurred())
		Expect(decoded).To(Equal(password))
	})

	It("sets required query parameters", func() {
		q := parsed.Query()

		Expect(q.Get("sslmode")).To(Equal("require"))
		Expect(q.Get("search_path")).To(Equal("public"))
		Expect(q.Get("application_name")).To(Equal("my-app"))
		Expect(q.Get("pool_max_conns")).To(Equal("20"))
		Expect(q.Get("pool_min_conns")).To(Equal("5"))
		Expect(q.Get("pool_max_conn_lifetime")).To(Equal("30m0s"))
		Expect(q.Get("pool_max_conn_idle_time")).To(Equal("5m0s"))
		Expect(q.Get("pool_health_check_period")).To(Equal("1m0s"))
	})

	Context("when PreferSimpleProtocol is enabled", func() {
		BeforeEach(func() {
			config.PreferSimpleProtocol = true

			connStr = buildConnectionString(
				host,
				user,
				password,
				dbname,
				port,
				schema,
				sslmode,
				config,
			)

			var err error
			parsed, err = url.Parse(connStr)
			Expect(err).NotTo(HaveOccurred())
		})

		It("adds prefer_simple_protocol=true", func() {
			Expect(parsed.Query().Get("prefer_simple_protocol")).To(Equal("true"))
		})
	})

	Context("when PreferSimpleProtocol is disabled", func() {
		It("does not include prefer_simple_protocol", func() {
			Expect(parsed.Query().Has("prefer_simple_protocol")).To(BeFalse())
		})
	})

	Context("with different schemas", func() {
		DescribeTable(
			"sets search_path correctly",
			func(schema string) {
				connStr = buildConnectionString(
					host,
					user,
					password,
					dbname,
					port,
					schema,
					sslmode,
					config,
				)

				parsed, err := url.Parse(connStr)
				Expect(err).NotTo(HaveOccurred())
				Expect(parsed.Query().Get("search_path")).To(Equal(schema))
			},
			Entry("public schema", "public"),
			Entry("custom schema", "tenant_42"),
			Entry("multiple schemas", "public,extensions"),
		)
	})

	Context("with complex usernames and passwords", func() {
		type credentials struct {
			username string
			password string
		}

		DescribeTable(
			"round-trips credentials correctly",
			func(creds credentials) {
				connStr := buildConnectionString(
					host,
					creds.username,
					creds.password,
					dbname,
					port,
					schema,
					sslmode,
					config,
				)

				parsed, err := url.Parse(connStr)
				Expect(err).NotTo(HaveOccurred())

				// Username is not escaped by net/url — should be exact
				Expect(parsed.User.Username()).To(Equal(creds.username))

				pw, ok := parsed.User.Password()
				Expect(ok).To(BeTrue())

				decoded, err := url.PathUnescape(pw)
				Expect(err).NotTo(HaveOccurred())
				Expect(decoded).To(Equal(creds.password))
			},
			Entry("simple credentials", credentials{
				username: "user",
				password: "password",
			}),
			Entry("password with URL reserved characters", credentials{
				username: "user",
				password: "p@ss:/?#[]@!$&'()*+,;=",
			}),
			Entry("password with spaces", credentials{
				username: "user",
				password: "pass word with spaces",
			}),
			Entry("password with unicode characters", credentials{
				username: "user",
				password: "pässwörd🔥雪",
			}),
			Entry("username with dots and dashes", credentials{
				username: "user.name-test",
				password: "password",
			}),
			Entry("username with spaces", credentials{
				username: "user name",
				password: "password",
			}),
			Entry("username and password both complex", credentials{
				username: "u$er:na/me?",
				password: "p@ss/w:rd?&=#+",
			}),
			Entry("credentials with leading and trailing whitespace", credentials{
				username: " user ",
				password: " password ",
			}),
			Entry("credentials with newline and tab characters", credentials{
				username: "user\tname",
				password: "pass\nword",
			}),
			Entry("credentials with emoji and symbols", credentials{
				username: "👩‍💻dev",
				password: "🔐pa$$🚀",
			}),
		)
	})
})
