import yaml

import click


@click.group()
def chart():
    pass


@click.argument('value', type=str)
@click.argument('output', type=click.File('w'))
@click.argument('input', type=click.File('rb'))
@click.command()
@click.pass_context
def write(ctx, input, output, value):
    data = yaml.load(input, Loader=yaml.FullLoader)
    data['image']['repository'] = value
    yaml.dump(data, output)


chart.add_command(write)


if __name__ == '__main__':
    chart()
