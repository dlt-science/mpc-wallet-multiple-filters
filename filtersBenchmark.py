import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns

# File paths
btc_bloom_path = 'BitcoinBloomResults.csv'
eth_cuckoo_path = 'EthereumCuckooResults.csv'
eth_bloom_path = 'EthereumBloomResults.csv'
btc_cuckoo_path = 'BitcoinCuckooResults.csv'

# Reading the data
btc_bloom = pd.read_csv(btc_bloom_path)
eth_cuckoo = pd.read_csv(eth_cuckoo_path)
eth_bloom = pd.read_csv(eth_bloom_path)
btc_cuckoo = pd.read_csv(btc_cuckoo_path)

# Combining Ethereum and Bitcoin data for easier plotting
eth_data = pd.concat([eth_bloom, eth_cuckoo], keys=['Bloom', 'Cuckoo'])
btc_data = pd.concat([btc_bloom, btc_cuckoo], keys=['Bloom', 'Cuckoo'])

# Adding 'Filter' and 'Blockchain' columns for clarity
eth_data['Filter'] = eth_data.index.get_level_values(0)
eth_data['Blockchain'] = 'Ethereum'
btc_data['Filter'] = btc_data.index.get_level_values(0)
btc_data['Blockchain'] = 'Bitcoin'

# Combined data
combined_data = pd.concat([eth_data, btc_data])

# # Adjusting Ethereum transaction fee from Wei to Ether for consistency
# combined_data['transactionFeeEther'] = combined_data.apply(
#     lambda row: row['transactionFeeWei'] * 1e-18 if row['Blockchain'] == 'Ethereum' else row['transactionFeeBTC'],
#     axis=1
# )

# Plot settings for readability on A4 paper
# plt.rcParams.update({'font.size': 10})

# Set the style of the plots to be 'whitegrid'
sns.set_style('white')

# Set the context to 'paper' which is suited for smaller plots
sns.set_context('paper')

# Set the color palette
# sns.color_palette("tab10")
sns.set_palette(["#3498db", "#e67e22"])


# Function to create the plots
def create_plot(data, y_column, y_label, filename):
    # Create a figure
    plt.figure()
    # plt.figure(figsize=(10, 6))
    # sns.barplot(x='partySet', y=y_column, hue='Filter', data=data, palette=['blue', 'green'])
    sns.barplot(x='partySet', y=y_column, hue='Filter', data=data)

    # Set the font size for the x and y axis labels
    plt.xlabel('Multi-signature wallet configuration', fontsize=18)
    plt.ylabel(y_label, fontsize=18)

    # Remove the frame
    sns.despine()

    # Increase the font size of the ticks
    plt.tick_params(axis='both', labelsize=16)

    # Removing the frame
    plt.legend(loc='upper center', bbox_to_anchor=(0.5, 1.1), ncol=2, frameon=False, fontsize=16)
    # plt.legend(title='Filters', bbox_to_anchor=(1.05, 1), loc='upper left')
    # plt.xticks(rotation=45)
    plt.tight_layout()
    plt.savefig(filename, dpi=300)


# Creating the four plots
create_plot(eth_data, 'transactionSize', 'Transaction size (bytes)',
            'eth_transaction_size_plot.pdf')
create_plot(btc_data, 'transactionSize', 'Transaction size (bytes)',
            'btc_transaction_size_plot.pdf')
create_plot(eth_data, 'transactionFeeEther', 'Transaction Fee (Ether)',
            'eth_transaction_fee_plot.pdf')
create_plot(btc_data, 'transactionFeeBTC', 'Transaction Fee (BTC)',
            'btc_transaction_fee_plot.pdf')
