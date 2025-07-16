import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'Unified OpenAI-Compatible API',
    Svg: require('@site/static/img/undraw_docusaurus_mountain.svg').default,
    description: (
      <>
        Exposes a single OpenAI-compatible `/v1/chat/completions` endpoint for seamless integration with various LLM providers.
      </>
    ),
  },
  {
    title: 'Multiple Provider Support',
    Svg: require('@site/static/img/undraw_docusaurus_tree.svg').default,
    description: (
      <>
        Supports OpenAI, Google Gemini, Ollama, and a Dummy Provider, allowing you to switch between LLMs effortlessly.
      </>
    ),
  },
  {
    title: 'Comprehensive Monitoring',
    Svg: require('@site/static/img/undraw_docusaurus_react.svg').default,
    description: (
      <>
        Integrates with Prometheus for token usage metrics and includes a pre-configured Grafana dashboard for visualization.
      </>
    ),
  },
  {
    title: 'Flexible Configuration',
    Svg: require('@site/static/img/undraw_docusaurus_mountain.svg').default,
    description: (
      <>
        Configurable via a YAML file and environment variables, providing flexibility for different deployment scenarios.
      </>
    ),
  },
  {
    title: 'OpenAPI Specification & Swagger UI',
    Svg: require('@site/static/img/undraw_docusaurus_tree.svg').default,
    description: (
      <>
        Provides an OpenAPI 3.0 specification and an interactive Swagger UI for easy API exploration and testing.
      </>
    ),
  },
  {
    title: 'Docker Support',
    Svg: require('@site/static/img/undraw_docusaurus_react.svg').default,
    description: (
      <>
        Easily deployable using Docker and Docker Compose for quick setup and consistent environments.
      </>
    ),
  },
];

function Feature({Svg, title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        }
        </div>
      </div>
    </section>
  );
}